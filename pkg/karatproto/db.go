package karatproto

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	//_ "github.com/golang-migrate/migrate/v4/source/file"
	//	database "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/go_bindata"
	st "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	sqlx "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ws-lab/karat-proto/pkg/migrations"
	"github.com/ws-lab/karat-proto/pkg/pb"
)

var (
	db            *sqlx.DB
	conf          = make(map[int]ConfStruct)
	BASE_PATH     string
	SYSTEM_PATH   string
	error_packet  = make(map[string]map[string][]byte)
	error_archive = make(map[string]map[int]FlagStruct)
	event_flag    = make(map[int]FlagStruct)
	con_flag      = make(map[int]FlagStruct)
	lora_flag     = make(map[int]FlagStruct)
	arch_flag     = make(map[int]FlagStruct)
	arch_flag2    = make(map[int]int) //ext_id  - bit
	log_flag      = make(map[int]FlagStruct)
)

func InitDB() {
	var err error
	path_db := SYSTEM_PATH + "/database"
	if _, err := os.Stat(path_db); os.IsNotExist(err) {
		os.Mkdir(path_db, 0777)
	}
	if err != nil {
		log.Fatal(err)
	}
	db, err = sqlx.Open("sqlite3", BASE_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			return
		}
	}()
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	migrationDB()
	log.Println("DataBase ready...")
	if err = setMapConfPacket(db); err != nil {
		log.Fatal(err)
	}
	log.Println("Conf Packet ready...")
	if err = setMapErrorPacket(db); err != nil {
		log.Fatal(err)
	}
	//Получение номери бита из байта ошибки и статуса
	if err = updateFlags(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapErrorArchive(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapEventFlag(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapConFlag(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapLoraFlag(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapArchFlag(db); err != nil {
		log.Fatal(err)
	}
	if err = setMapLogCode(db); err != nil {
		log.Fatal(err)
	}
	log.Println("Data validate...")
}
func setMapConfPacket(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select de.`code`,de.`index` as `index`,rv.`resource_id`,de.`rvalue_id`,de.len,de.type_data,de.little_endian,de.is_sensor,r.mask_type,d.vr,d.`dmodels`,coalesce(rv.name,r.alias) alias,d.packet_type,(select multiplier from unit u where u.id=coalesce(de.unit_id,rv.unit_id)) multiplier from dconf_ext as de join dconf as d on d.id=de.code join rvariable as r on de.rvariable_id=r.id left join rvalue rv on de.rvalue_id = rv.id order by de.`index`")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	conf_ := make(map[int][]Scan)
	conf_version := make(map[int]string)
	conf_models := make(map[int][]string)
	conf_packet_type := make(map[int]int)
	conf = make(map[int]ConfStruct)
	scan := Scan{}

	for rows.Next() {
		if err = rows.StructScan(&scan); err != nil {
			log.Println(err)
			return err
		}
		if _, ok := conf_[scan.Code]; !ok {
			conf_[scan.Code] = []Scan{}
		}
		if _, ok := conf[scan.Code]; !ok {
			conf[scan.Code] = ConfStruct{Scans: []Scan{}}
		}
		conf_[scan.Code] = append(conf_[scan.Code], scan)
		if scan.VR == "" {
			log.Fatal(errors.New("У пакета нет данных о версии протокола!!"))
		}
		if _, ok := conf_version[scan.Code]; !ok {
			conf_version[scan.Code] = scan.VR
		}
		if _, ok := conf_models[scan.Code]; !ok && scan.Models != "" {
			conf_models[scan.Code] = strings.Split(scan.Models, ",")
		}
		if _, ok := conf_packet_type[scan.Code]; !ok {
			conf_packet_type[scan.Code] = scan.PacketType
		}
	}
	for i, _ := range conf {
		conf[i] = ConfStruct{Scans: conf_[i], ConfVersion: conf_version[i], ConfModel: conf_models[i], PacketType: conf_packet_type[i]}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
func setMapErrorPacket(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select vr,note,byte from packet_error order by vr,byte")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	error_packet = make(map[string]map[string][]byte)
	var (
		vr     string
		note   string
		byte_  string
		vr_arr []string
	)

	for rows.Next() {
		if err = rows.Scan(&vr, &note, &byte_); err != nil {
			return err
		}
		vr_arr = strings.Split(vr, ",")
		for _, item := range vr_arr {
			if _, ok := error_packet[item]; !ok {
				error_packet[item] = make(map[string][]byte)
			}
			b, err := getByte(byte_)
			if err != nil {
				return err
			}
			error_packet[item][note] = b
		}

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
func setMapErrorArchive(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,note,device,nbit,byte from flags where type=1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	error_archive = make(map[string]map[int]FlagStruct)
	var (
		id         int
		note       string
		device     string
		device_arr []string
		nbit       int
		byte       string
	)

	for rows.Next() {
		if err = rows.Scan(&id, &note, &device, &nbit, &byte); err != nil {
			return err
		}
		device_arr = strings.Split(device, ",")
		for _, item := range device_arr {
			if _, ok := error_archive[item]; !ok {
				error_archive[item] = make(map[int]FlagStruct)
			}
			error_archive[item][nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
		}

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
func setMapEventFlag(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,note,nbit,byte from flags where type=2")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	event_flag = make(map[int]FlagStruct)
	var (
		id   int
		note string
		nbit int
		byte string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &note, &nbit, &byte); err != nil {
			return err
		}
		event_flag[nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
func setMapConFlag(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,note,nbit,byte from flags where type=3")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	con_flag = make(map[int]FlagStruct)
	var (
		id   int
		note string
		nbit int
		byte string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &note, &nbit, &byte); err != nil {
			return err
		}
		con_flag[nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}

//Флаги для функции 0x0021
func setMapLoraFlag(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,note,nbit,byte from flags where type=4")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	lora_flag = make(map[int]FlagStruct)
	var (
		id   int
		note string
		nbit int
		byte string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &note, &nbit, &byte); err != nil {
			return err
		}
		lora_flag[nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}

//Флаги архивов
func setMapArchFlag(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select ext_id,note,nbit,byte from flags where type=5")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	arch_flag = make(map[int]FlagStruct)
	var (
		id   int
		note string
		nbit int
		byte string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &note, &nbit, &byte); err != nil {
			return err
		}
		arch_flag[nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
		arch_flag2[id] = nbit
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}

//Флаги нештатных ситуаций
func setMapLogCode(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,note,nbit,byte from flags where type in (2,6)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	log_flag = make(map[int]FlagStruct)
	var (
		id   int
		note string
		nbit int
		byte string
	)
	for rows.Next() {
		if err = rows.Scan(&id, &note, &nbit, &byte); err != nil {
			return err
		}
		log_flag[nbit] = FlagStruct{Id: id, Note: note, Byte: byte}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return
}
func updateFlags(db *sqlx.DB) (err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,byte from flags where /*nbit ='' and*/ byte != ''")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var (
		id    int
		byte_ string
	)
	tx := db.MustBegin()
	for rows.Next() {
		if err = rows.Scan(&id, &byte_); err != nil {
			return err
		}
		b, err := getByte(byte_)
		if err != nil {
			return err
		}
		val := getFlagStatusNByteTemp(b)
		_, err = tx.NamedExec("update flags set nbit = :nbit where id=:id",
			map[string]interface{}{
				"nbit": val,
				"id":   id,
			})
		if err != nil {
			log.Fatalln(err)
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
	return
}
func getPacket(db *sqlx.DB) (packet *pb.PacketResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,vr,dmodels,packet_type from dconf")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.Packet, 0)
	for rows.Next() {
		var rec Packet
		var rec_pb pb.Packet
		if err = rows.StructScan(&rec); err != nil {
			log.Println(err)
			return packet, err
		}
		rec_pb = pb.Packet{Id: rec.Id, Vr: rec.Vr, PacketType: rec.PacketType, Models: strings.Split(rec.Models, ",")}
		rs = append(rs, &rec_pb)
	}
	packet = &pb.PacketResponse{Packet: rs}
	return
}

func getPacketConf(db *sqlx.DB, code int32) (packet *pb.PacketConfResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	sql := "select id,code,`index`,resource_id,rvariable_id,rvalue_id,`len`,type_data,little_endian,is_sensor,unit_id from dconf_ext"
	if code > 0 {
		sql = sql + " where code = " + strconv.Itoa(int(code))
	}
	rows, err := db.Queryx(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.PacketConf, 0)
	for rows.Next() {
		var rec PacketConf
		var rec_pb pb.PacketConf
		if err = rows.StructScan(&rec); err != nil {
			log.Println(err)
			return packet, err
		}
		var rvalue_id *st.Value
		if rec.RvalueId.Valid {
			rvalue_id = &st.Value{
				Kind: &st.Value_NumberValue{
					NumberValue: float64(rec.RvalueId.Int64),
				},
			}
		} else {
			rvalue_id = &st.Value{
				Kind: &st.Value_NullValue{},
			}
		}
		var unit_id *st.Value
		if rec.UnitId.Valid {
			unit_id = &st.Value{
				Kind: &st.Value_NumberValue{
					NumberValue: float64(rec.UnitId.Int64),
				},
			}
		} else {
			unit_id = &st.Value{
				Kind: &st.Value_NullValue{},
			}
		}
		rec_pb = pb.PacketConf{Id: rec.Id, Code: rec.Code, Index: rec.Index, TypeData: rec.TypeData, ResourceId: rec.ResourceId, RvalueId: rvalue_id, RvariableId: rec.RvariableId, Len: rec.Len, LittleEndian: &wrappers.BoolValue{Value: rec.LittleEndian}, IsSensor: &wrappers.BoolValue{Value: rec.IsSensor}, UnitId: unit_id}
		rs = append(rs, &rec_pb)
	}
	packet = &pb.PacketConfResponse{PacketConf: rs}
	return
}

func getResource(db *sqlx.DB) (resources *pb.ResourceResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,name from resource")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.Resource, 0)
	for rows.Next() {
		var rec pb.Resource
		if err = rows.StructScan(&rec); err != nil {
			log.Println(err)
			return resources, err
		}
		rs = append(rs, &rec)
	}
	resources = &pb.ResourceResponse{Resource: rs}
	return
}
func getRvariable(db *sqlx.DB) (rvariables *pb.RvariableResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
	}
	rows, err := db.Queryx("select id,name,alias from rvariable")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.Rvariable, 0)
	for rows.Next() {
		var rec pb.Rvariable
		if err = rows.StructScan(&rec); err != nil {
			log.Println(err)
			return rvariables, err
		}
		rs = append(rs, &rec)
	}
	rvariables = &pb.RvariableResponse{Rvariable: rs}
	return
}
func getRvalue(db *sqlx.DB) (rvalues *pb.RvalueResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := db.Queryx("select `id`,name,`resource_id`,`rvariable_id`,`unit_id` from rvalue")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.Rvalue, 0)
	for rows.Next() {
		var rec Rvalue
		var rec_pb pb.Rvalue
		if err = rows.Scan(&rec.Id, &rec.Name, &rec.ResourceId, &rec.RvariableId, &rec.UnitId); err != nil {
			log.Println(err)
			return rvalues, err
		}
		var unit_id *st.Value
		if rec.UnitId.Valid {
			unit_id = &st.Value{
				Kind: &st.Value_NumberValue{
					NumberValue: float64(rec.UnitId.Int64),
				},
			}
		} else {
			unit_id = &st.Value{
				Kind: &st.Value_NullValue{},
			}
		}
		rec_pb = pb.Rvalue{Id: rec.Id, Name: rec.Name, RvariableId: rec.RvariableId, ResourceId: rec.ResourceId, UnitId: unit_id}
		rs = append(rs, &rec_pb)
	}
	rvalues = &pb.RvalueResponse{Rvalues: rs}
	return
}
func getUnit(db *sqlx.DB) (units *pb.UnitResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := db.Queryx("select `id`,`rvariable_id`,name,multiplier from unit")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.Unit, 0)
	for rows.Next() {
		var rec pb.Unit
		if err = rows.Scan(&rec.Id, &rec.RvariableId, &rec.Name, &rec.Multiplier); err != nil {
			log.Println(err)
			return units, err
		}
		rs = append(rs, &rec)
	}
	units = &pb.UnitResponse{Units: rs}
	return
}
func setSettingsPacketProtocol20(db *sqlx.DB, req *pb.SettingsPacketProtocol20Request) (ret *pb.SettingsPacketProtocol20Response, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
		if err != nil {
			log.Fatal(err)
		}
	}
	var ip SettingsPacketByteTX20
	var event_arr [16]int
	if len(req.GetEventMaskFlags()) > 0 {
		var i int
		for _, item_req := range req.GetEventMaskFlags() {
			for index, item := range event_flag {
				if item_req == int32(item.Id) {
					i++
					event_arr[index] = 1
				}
			}
		}
		if i != len(req.GetEventMaskFlags()) {
			return ret, errors.New("Передано не известное значение в параметре event_mask_flags")
		}
		for i, j := 0, len(event_arr)-1; i < j; i, j = i+1, j-1 {
			event_arr[i], event_arr[j] = event_arr[j], event_arr[i]
		}
		var event_bits string
		for _, aa := range event_arr {
			event_bits += strconv.Itoa(aa)
		}
		//log.Println(event_bits)
		var eb int64
		if eb, err = strconv.ParseInt(event_bits, 2, 64); err != nil {
			log.Println(err)
			return
		}
		ebb := make([]byte, 2)
		binary.LittleEndian.PutUint16(ebb, uint16(eb))
		//log.Printf("%+X", ebb)
		copy(ip.EventMaskFlags[:], ebb)
	}
	abb, err := setArchByte(req.GetArchFlags())
	if err != nil {
		log.Println(err)
		return
	}
	copy(ip.ArchType[:], abb)

	if err := ReflectStructField(req, "DeltaDevTime"); err == nil {
		lwtime := make([]byte, 4)
		binary.LittleEndian.PutUint32(lwtime, uint32(req.GetDeltaDevTime()))
		copy(ip.DeltaDevTime[:], lwtime)
	} else {
		return ret, errors.New("Не задан обязательный параметр dev_time")
	}
	if err := ReflectStructField(req, "TimeZone"); err == nil {
		ip.TimeZone = int8(req.GetTimeZone() * 60 / 15)
	} else {
		return ret, errors.New("Не задан обязательный параметр time_zone")
	}

	if err := ReflectStructField(req, "TxPeriod"); err == nil && req.GetTxPeriod() <= 720 && req.GetTxPeriod() > 0 {
		tx_period := make([]byte, 4)
		binary.LittleEndian.PutUint32(tx_period, uint32(req.GetTxPeriod()))
		copy(ip.TxPeriod[:], tx_period)
	} else {
		return ret, errors.New("Не задан обязательный параметр tx_period или задан не верно")
	}
	if err := ReflectStructField(req, "OffsetTime"); err == nil && req.GetOffsetTime() <= 43200 && req.GetOffsetTime() > 0 {
		offset_time := make([]byte, 4)
		binary.LittleEndian.PutUint16(offset_time, uint16(req.GetOffsetTime()))
		copy(ip.OffsetTime[:], offset_time)
	} else {
		return ret, errors.New("Не задан обязательный параметр offset_time или задан не верно")
	}

	code_function := make([]byte, 2)
	binary.LittleEndian.PutUint16(code_function, uint16(257))
	copy(ip.CodeFunction[:], code_function)

	if err := ReflectStructField(req, "MainMsgCnt"); err == nil && req.GetMainMsgCnt() > 0 {
		ip.MainMsgCnt = uint8(req.GetMainMsgCnt())
	} else {
		return ret, errors.New("Не задан обязательный параметр main_msg_cnt или задан не верно")
	}

	if err := ReflectStructField(req, "Mmsg1"); err == nil && req.GetMmsg1() > 0 {
		msg1 := make([]byte, 2)
		binary.LittleEndian.PutUint16(msg1, uint16(req.GetMmsg1()))
		copy(ip.MMsg1[:], msg1)
	} else {
		return ret, errors.New("Не задан обязательный параметр mmsg1 или задан не верно")
	}
	if err := ReflectStructField(req, "Mmsg2"); (err != nil || req.GetMmsg2() == 0) && req.GetMainMsgCnt() > 1 {
		return ret, errors.New("Не задан параметр mmsg2 или задан не верно")
	}
	msg2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg2, uint16(req.GetMmsg2()))
	copy(ip.MMsg2[:], msg2)

	if err := ReflectStructField(req, "Mmsg3"); (err != nil || req.GetMmsg3() == 0) && req.GetMainMsgCnt() > 2 {
		return ret, errors.New("Не задан параметр mmsg3 или задан не верно")
	}
	msg3 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg3, uint16(req.GetMmsg3()))
	copy(ip.MMsg3[:], msg3)

	if err := ReflectStructField(req, "Mmsg4"); (err != nil || req.GetMmsg4() == 0) && req.GetMainMsgCnt() > 3 {
		return ret, errors.New("Не задан параметр mmsg4 или задан не верно")
	}
	msg4 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg4, uint16(req.GetMmsg4()))
	copy(ip.MMsg4[:], msg4)

	if err := ReflectStructField(req, "Mmsg5"); (err != nil || req.GetMmsg5() == 0) && req.GetMainMsgCnt() > 4 {
		return ret, errors.New("Не задан параметр mmsg5 или задан не верно")
	}
	msg5 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg5, uint16(req.GetMmsg5()))
	copy(ip.MMsg5[:], msg5)

	if err := ReflectStructField(req, "Mmsg6"); (err != nil || req.GetMmsg6() == 0) && req.GetMainMsgCnt() > 5 {
		return ret, errors.New("Не задан параметр mmsg6 или задан не верно")
	}
	msg6 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg6, uint16(req.GetMmsg6()))
	copy(ip.MMsg6[:], msg6)

	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, ip)
	ret = &pb.SettingsPacketProtocol20Response{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}
func setSettingsPacketProtocol18(db *sqlx.DB, req *pb.SettingsPacketProtocol18Request) (ret *pb.SettingsPacketProtocol18Response, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
		if err != nil {
			log.Fatal(err)
		}
	}
	var ip SettingsPacketByteTX18
	abb, err := setArchByte(req.GetArchFlags())
	if err != nil {
		log.Println(err)
		return
	}
	copy(ip.ArchType[:], abb)

	if err := ReflectStructField(req, "DeltaDevTime"); err != nil {
		if err := ReflectStructField(req, "LwTime"); err != nil && req.GetLwTime() == "" {
			return ret, errors.New("Не задан обязательный параметр lw_time/delta_dev_time")
		}
	}

	if err := ReflectStructField(req, "DeltaDevTime"); err == nil {
		if err := ReflectStructField(req, "LwTime"); err == nil && req.GetDeltaDevTime() > 0 && req.GetLwTime() != "" {
			return ret, errors.New("Нельзя задать одновременно и lw_time и delta_dev_time")
		}
		delta_dev_time := make([]byte, 4)
		binary.LittleEndian.PutUint32(delta_dev_time, uint32(req.GetDeltaDevTime()))
		copy(ip.DeltaDevTime[:], delta_dev_time)

		ip.LWTime = LWTime{Sec: uint8(255), Min: uint8(255), Hour: uint8(255), Day: uint8(255), Month: uint8(255), Year: uint8(255)}
	}

	if err := ReflectStructField(req, "LwTime"); err == nil && req.GetLwTime() != "" {
		format := "2006-01-02 15:04:05"
		tt, err := time.Parse(format, req.GetLwTime())
		if err != nil {
			return ret, err
		}
		ip.LWTime = LWTime{Sec: uint8(tt.Second()), Min: uint8(tt.Minute()), Hour: uint8(tt.Hour()), Day: uint8(tt.Day()), Month: uint8(tt.Month()), Year: uint8((tt.Year() - 2000))}

		delta_dev_time := make([]byte, 4)
		binary.LittleEndian.PutUint32(delta_dev_time, uint32(0))
		copy(ip.DeltaDevTime[:], delta_dev_time)
	}

	if err := ReflectStructField(req, "TimeZone"); err == nil {
		ip.TimeZone = uint8(req.GetTimeZone() * 60 / 15)
	} else {
		return ret, errors.New("Не задан обязательный параметр time_zone")
	}

	if err := ReflectStructField(req, "TxPeriod"); err == nil /*&& req.GetTxPeriod() <= 720*/ && req.GetTxPeriod() > 0 {
		tx_period := make([]byte, 4)
		binary.LittleEndian.PutUint32(tx_period, uint32(req.GetTxPeriod()))
		copy(ip.TxPeriod[:], tx_period)
	} else {
		return ret, errors.New("Не задан обязательный параметр tx_period или задан не верно")
	}
	if err := ReflectStructField(req, "OffsetTime"); err == nil /*&& req.GetOffsetTime() <= 43200*/ && req.GetOffsetTime() >= 0 {
		offset_time := make([]byte, 4)
		binary.LittleEndian.PutUint16(offset_time, uint16(req.GetOffsetTime()))
		copy(ip.OffsetTime[:], offset_time)
	} else {
		return ret, errors.New("Не задан обязательный параметр offset_time или задан не верно")
	}

	code_function := make([]byte, 2)
	binary.LittleEndian.PutUint16(code_function, uint16(21))
	copy(ip.CodeFunction[:], code_function)

	if err := ReflectStructField(req, "MainMsgCnt"); err == nil && req.GetMainMsgCnt() > 0 {
		ip.MainMsgCnt = uint8(req.GetMainMsgCnt())
	} else {
		return ret, errors.New("Не задан обязательный параметр main_msg_cnt или задан не верно")
	}

	if err := ReflectStructField(req, "Mmsg1"); err == nil && req.GetMmsg1() > 0 {
		msg1 := make([]byte, 2)
		binary.LittleEndian.PutUint16(msg1, uint16(req.GetMmsg1()))
		copy(ip.MMsg1[:], msg1)
	} else {
		return ret, errors.New("Не задан обязательный параметр mmsg1 или задан не верно")
	}
	if err := ReflectStructField(req, "Mmsg2"); (err != nil || req.GetMmsg2() == 0) && req.GetMainMsgCnt() > 1 {
		return ret, errors.New("Не задан параметр mmsg2 или задан не верно")
	}
	msg2 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg2, uint16(req.GetMmsg2()))
	copy(ip.MMsg2[:], msg2)

	if err := ReflectStructField(req, "Mmsg3"); (err != nil || req.GetMmsg3() == 0) && req.GetMainMsgCnt() > 2 {
		return ret, errors.New("Не задан параметр mmsg3 или задан не верно")
	}
	msg3 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg3, uint16(req.GetMmsg3()))
	copy(ip.MMsg3[:], msg3)

	if err := ReflectStructField(req, "Mmsg4"); (err != nil || req.GetMmsg4() == 0) && req.GetMainMsgCnt() > 3 {
		return ret, errors.New("Не задан параметр mmsg4 или задан не верно")
	}
	msg4 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg4, uint16(req.GetMmsg4()))
	copy(ip.MMsg4[:], msg4)

	if err := ReflectStructField(req, "Mmsg5"); (err != nil || req.GetMmsg5() == 0) && req.GetMainMsgCnt() > 4 {
		return ret, errors.New("Не задан параметр mmsg5 или задан не верно")
	}
	msg5 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg5, uint16(req.GetMmsg5()))
	copy(ip.MMsg5[:], msg5)

	if err := ReflectStructField(req, "Mmsg6"); (err != nil || req.GetMmsg6() == 0) && req.GetMainMsgCnt() > 5 {
		return ret, errors.New("Не задан параметр mmsg6 или задан не верно")
	}
	msg6 := make([]byte, 2)
	binary.LittleEndian.PutUint16(msg6, uint16(req.GetMmsg6()))
	copy(ip.MMsg6[:], msg6)

	if err := ReflectStructField(req, "RepDate"); err != nil {
		return ret, errors.New("Не задан параметр rep_date или задан не верно")
	}
	ip.BegPeriod = uint8(req.GetRepDate())

	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, ip)
	ret = &pb.SettingsPacketProtocol18Response{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}

func setTimeCorrect(req *pb.TimeCorrectRequest) (ret *pb.TimeCorrectResponse, err error) {
	var ip TimeCorrectByteTx
	code_function := make([]byte, 2)
	binary.LittleEndian.PutUint16(code_function, uint16(22))
	copy(ip.CodeFunction[:], code_function)

	if err := ReflectStructField(req, "DeltaTime"); err == nil && req.GetDeltaTime() != 0 {
		delta_time := make([]byte, 4)
		binary.LittleEndian.PutUint32(delta_time, uint32(req.GetDeltaTime()))
		copy(ip.DeltaDevTime[:], delta_time)
	} else {
		return ret, errors.New("Не задан параметр delta_time или задан не верно")
	}
	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, ip)
	ret = &pb.TimeCorrectResponse{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}

func setArchQuery(req *pb.ArchQueryRequest) (ret *pb.ArchQueryResponse, err error) {
	var bin_buf bytes.Buffer
	if _, ok := arch_flag2[int(req.GetArchType())]; !ok {
		return ret, errors.New("Не задан параметр arch_type или задан не верно")
	}
	if err := ReflectStructField(req, "ArchTime"); err != nil || req.GetArchTime() == "" {
		return ret, errors.New("Не задан параметр arch_time или задан не верно")
	}
	if err := ReflectStructField(req, "ProtocolVersion"); err == nil && req.GetProtocolVersion().String() != "" {
		code_function := make([]byte, 2)
		format := "2006-01-02 15:04:05"
		tt, err := time.Parse(format, req.GetArchTime())
		if err != nil {
			return ret, err
		}
		if req.GetProtocolVersion().String() == "PROTOCOL20" {
			var ip setArchivePacketByteTx20
			ip.ArchType = uint8(req.GetArchType())
			arch_time := int(Second2LwTimeExtWRITE2(tt))
			binary.LittleEndian.PutUint16(code_function, uint16(58))
			copy(ip.CodeFunction[:], code_function)
			arch_time_byte := make([]byte, 4)
			binary.LittleEndian.PutUint32(arch_time_byte, uint32(arch_time))
			copy(ip.ArchTime[:], arch_time_byte)
			binary.Write(&bin_buf, binary.BigEndian, ip)
		} else if req.GetProtocolVersion().String() == "PROTOCOL18" {
			var ip setArchivePacketByteTx18
			ip.ArchType = uint8(req.GetArchType())
			binary.LittleEndian.PutUint16(code_function, uint16(12))
			copy(ip.CodeFunction[:], code_function)
			ip.Hour = uint8(tt.Hour())
			ip.Day = uint8(tt.Day())
			ip.Month = uint8(tt.Month())
			ip.Year = uint8((tt.Year() - 2000))
			binary.Write(&bin_buf, binary.BigEndian, ip)
		} else {
			return ret, errors.New("Не задан параметр protocol_version или задан не верно")
		}

	} else {
		return ret, errors.New("Не задан параметр protocol_version или задан не верно")
	}

	ret = &pb.ArchQueryResponse{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}
func setLoraWanParams(req *pb.LoraWanParamsRequest) (ret *pb.LoraWanParamsResponse, err error) {
	var ip setLoraWanParamsByteTx20
	code_function := make([]byte, 2)
	binary.LittleEndian.PutUint16(code_function, uint16(33))
	copy(ip.CodeFunction[:], code_function)

	if err := ReflectStructField(req, "JoinParams"); err == nil && req.GetJoinParams() > 0 {
		ip.JoinParams = uint8(req.GetJoinParams())
	} else {
		return ret, errors.New("Не задан параметр join_params или задан не верно")
	}
	if err := ReflectStructField(req, "Dn2Freq"); err == nil && req.GetDn2Freq() > 0 {
		dn2freq := make([]byte, 4)
		binary.LittleEndian.PutUint32(dn2freq, uint32(req.GetDn2Freq()))
		copy(ip.DN2Freq[:], dn2freq)
	} else {
		return ret, errors.New("Не задан параметр dn2freq или задан не верно")
	}
	if err := ReflectStructField(req, "Port"); err == nil && req.GetPort() > 0 {
		ip.Port = uint8(req.GetPort())
	} else {
		return ret, errors.New("Не задан параметр port или задан не верно")
	}
	if err := ReflectStructField(req, "TxAtts"); err == nil && req.GetTxAtts() > 0 {
		ip.TxAtts = uint8(req.GetTxAtts())
	} else {
		return ret, errors.New("Не задан параметр tx_atts или задан не верно")
	}
	if err := ReflectStructField(req, "AdrParam"); err == nil && req.GetAdrParam() > 0 {
		ip.ADRparam = uint8(req.GetAdrParam())
	} else {
		return ret, errors.New("Не задан параметр adr_param или задан не верно")
	}

	if len(req.GetCflags()) > 0 {
		var (
			i          int
			cflags_arr [8]int
		)
		for _, item_req := range req.GetCflags() {
			for index, item := range lora_flag {
				if item_req == int32(item.Id) {
					i++
					cflags_arr[index] = 1
				}
			}
		}
		if i != len(req.GetCflags()) {
			return ret, errors.New("Передано не известное значение в параметре cflags")
		}
		for i, j := 0, len(cflags_arr)-1; i < j; i, j = i+1, j-1 {
			cflags_arr[i], cflags_arr[j] = cflags_arr[j], cflags_arr[i]
		}
		var cflogs_bits string
		for _, aa := range cflags_arr {
			cflogs_bits += strconv.Itoa(aa)
		}
		var eb int64
		if eb, err = strconv.ParseInt(cflogs_bits, 2, 64); err != nil {
			log.Println(err)
			return
		}
		ebb := make([]byte, 2)
		binary.LittleEndian.PutUint16(ebb, uint16(eb))
		copy(ip.CFlags[:], ebb)
	}
	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, ip)
	ret = &pb.LoraWanParamsResponse{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}
func setFunctionWithEmptyParam(req *pb.FunctionWithEmptyParamRequest) (ret *pb.FunctionWithEmptyParamResponse, err error) {
	var CodeFunction [2]byte
	if err := ReflectStructField(req, "FunctionCode"); err == nil && req.GetFunctionCode() > 0 {
		function_code := make([]byte, 2)
		binary.LittleEndian.PutUint16(function_code, uint16(req.GetFunctionCode()))
		copy(CodeFunction[:], function_code)
	} else {
		return ret, errors.New("Не задан параметр function_code или задан не верно")
	}

	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, CodeFunction)
	ret = &pb.FunctionWithEmptyParamResponse{Spacket: fmt.Sprintf("%x", bin_buf.Bytes()), Bpacket: bin_buf.Bytes()}
	return
}
func Second2LwTimeExtWRITE(sec int64) (lw_time_diff float64) {
	now := time.Now().UTC()
	date := time.Date(2000, 03, 01, 0, 0, 0, 0, time.UTC)
	date = date.Add(time.Second * time.Duration(sec))
	lw_time_diff = now.Sub(date).Seconds()
	return
}
func Second2LwTimeExtWRITE2(datetime time.Time) (lw_time_diff float64) { //преобразование из времени в во время карата
	date := time.Date(2000, 03, 01, 0, 0, 0, 0, time.UTC)
	datetime = time.Date(datetime.Year(), datetime.Month(), datetime.Day(), datetime.Hour(), datetime.Minute(), datetime.Second(), 0, time.UTC)
	lw_time_diff = datetime.Sub(date).Seconds()
	return
}
func setArchByte(param []uint32) (abb []byte, err error) {
	var arch_arr [8]int
	if len(param) > 0 {
		for _, item_req := range param {
			if _, ok := arch_flag2[int(item_req)]; !ok {
				return abb, errors.New("Передано не известное значение в параметре arch_flags")
			}
			arch_arr[arch_flag2[int(item_req)]] = 1
		}
	} else {
		return abb, errors.New("Передано не известное значение в параметре arch_flags")
	}
	for i, j := 0, len(arch_arr)-1; i < j; i, j = i+1, j-1 {
		arch_arr[i], arch_arr[j] = arch_arr[j], arch_arr[i]
	}
	var archives_bits string
	for _, aa := range arch_arr {
		archives_bits += strconv.Itoa(aa)
	}
	var ab int64
	if ab, err = strconv.ParseInt(archives_bits, 2, 64); err != nil {
		return
	}
	abb = make([]byte, 2)
	binary.LittleEndian.PutUint16(abb, uint16(ab))
	return abb, err
}
func getFlagOfTypeList(db *sqlx.DB, type_id int) (archs *pb.FlagOfTypeResponse, err error) {
	err = db.Ping()
	if err != nil {
		db, err = sqlx.Open("sqlite3", BASE_PATH)
		if err != nil {
			log.Fatal(err)
		}
	}
	rows, err := db.Queryx("select case when type = 5 then ext_id else id end as id,note,byte,nbit,device from flags where type = $1", type_id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var rs = make([]*pb.FlagOfType, 0)
	for rows.Next() {
		var rec Flags
		var rec_pb pb.FlagOfType
		if err = rows.StructScan(&rec); err != nil {
			return archs, err
		}
		var nbit *st.Value
		nbit = &st.Value{
			Kind: &st.Value_NumberValue{
				NumberValue: float64(rec.Nbit),
			},
		}
		rec_pb = pb.FlagOfType{Id: rec.Id, Note: rec.Note, Byte: rec.Byte, Nbit: nbit, Devices: rec.Devices}
		rs = append(rs, &rec_pb)
	}
	archs = &pb.FlagOfTypeResponse{Flags: rs}
	return
}
func migrationDB() {
	db, err := sql.Open("sqlite3", BASE_PATH)
	if err != nil {
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			return
		}
	}()
	s := bindata.Resource(migrations.AssetNames(), func(name string) ([]byte, error) {
		return migrations.Asset(name)
	})
	d, _ := bindata.WithInstance(s)
	ss := &sqlite3.Sqlite{}
	ss.Open("sqlite3://" + BASE_PATH)
	m, err := migrate.NewWithSourceInstance(
		"go_bindata",
		d, "sqlite3://"+BASE_PATH)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		if err.Error() != "no change" {
			log.Fatal(err)
		}
	}
}
