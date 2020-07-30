package karatproto

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"
	"strings"
	s "strings"
	"time"

	"github.com/ws-lab/karat-proto/pkg/pb"
)

//GRPCServer
type GRPCServer struct{}

func (s *GRPCServer) PacketDecode(ctx context.Context, req *pb.PacketDecodeRequest) (*pb.PacketDecodeResponse, error) {
	var err error
	if req.GetBpacket() == nil && req.GetSpacket() == "" {
		return &pb.PacketDecodeResponse{}, nil
	}
	var sb []byte
	if req.GetSpacket() != "" {
		sb, err = str2Byte(req.GetSpacket())
		if err != nil {
			return nil, err
		}
	} else {
		sb, err = str2Byte(fmt.Sprintf("%+s", req.GetBpacket()))
		if err != nil {
			return nil, err
		}
	}
	func_code, resource_inx, error_type, err := getFuctionCode(sb)
	if err != nil {
		log.Println(err)
		return &pb.PacketDecodeResponse{}, err
	}

	error_const := ""
	if error_type {
		_, error_const_, err := getPacketError(sb, conf[int(func_code)].ConfVersion)
		if err != nil {
			return &pb.PacketDecodeResponse{Func: int32(func_code), PacketError: error_const, PacketType: int32(conf[int(func_code)].PacketType)}, err
		}
		error_const = error_const_
		return &pb.PacketDecodeResponse{Func: int32(func_code), ResourceInx: int32(resource_inx), PacketError: error_const, PacketType: int32(conf[int(func_code)].PacketType)}, nil
	}
	packet, err := packetDecode(int(func_code), sb)
	if err != nil {
		error_const = err.Error()
		return &pb.PacketDecodeResponse{Func: int32(func_code), PacketError: error_const, PacketType: int32(conf[int(func_code)].PacketType)}, err
	}
	return &pb.PacketDecodeResponse{Func: int32(func_code), Datas: packet, ResourceInx: int32(resource_inx), PacketError: error_const, PacketType: int32(conf[int(func_code)].PacketType)}, nil
}
func (s *GRPCServer) GetPacket(ctx context.Context, req *pb.EmptyRequest) (*pb.PacketResponse, error) {
	r, err := getPacket(db)
	if err != nil {
		return &pb.PacketResponse{}, err
	}
	return r, nil
}
func (s *GRPCServer) GetPacketConf(ctx context.Context, req *pb.PacketConfRequest) (*pb.PacketConfResponse, error) {
	packet_code := req.GetPacketCode()
	r, err := getPacketConf(db, packet_code)
	if err != nil {
		return &pb.PacketConfResponse{}, err
	}
	return r, nil
}
func (s *GRPCServer) GetResource(ctx context.Context, req *pb.EmptyRequest) (*pb.ResourceResponse, error) {
	r, err := getResource(db)
	if err != nil {
		return &pb.ResourceResponse{}, err
	}
	return r, nil
}

func (s *GRPCServer) GetRvariable(ctx context.Context, req *pb.EmptyRequest) (*pb.RvariableResponse, error) {
	r, err := getRvariable(db)
	if err != nil {
		return &pb.RvariableResponse{}, err
	}
	return r, nil
}

func (s *GRPCServer) GetRvalue(ctx context.Context, req *pb.EmptyRequest) (*pb.RvalueResponse, error) {
	r, err := getRvalue(db)
	if err != nil {
		return &pb.RvalueResponse{}, err
	}
	return r, err
}
func (s *GRPCServer) GetUnit(ctx context.Context, req *pb.EmptyRequest) (*pb.UnitResponse, error) {
	r, err := getUnit(db)
	if err != nil {
		return &pb.UnitResponse{}, err
	}
	return r, err
}

func (s *GRPCServer) SetSettingsPacketProtocol20(ctx context.Context, req *pb.SettingsPacketProtocol20Request) (ret *pb.SettingsPacketProtocol20Response, err error) {
	ret, err = setSettingsPacketProtocol20(db, req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func (s *GRPCServer) SetSettingsPacketProtocol18(ctx context.Context, req *pb.SettingsPacketProtocol18Request) (ret *pb.SettingsPacketProtocol18Response, err error) {
	ret, err = setSettingsPacketProtocol18(db, req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func (s *GRPCServer) SetTimeCorrect(ctx context.Context, req *pb.TimeCorrectRequest) (ret *pb.TimeCorrectResponse, err error) {
	ret, err = setTimeCorrect(req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func (s *GRPCServer) SetArchQuery(ctx context.Context, req *pb.ArchQueryRequest) (ret *pb.ArchQueryResponse, err error) {
	ret, err = setArchQuery(req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func (s *GRPCServer) SetLoraWanParams(ctx context.Context, req *pb.LoraWanParamsRequest) (ret *pb.LoraWanParamsResponse, err error) {
	ret, err = setLoraWanParams(req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func (s *GRPCServer) SetFunctionWithEmptyParam(ctx context.Context, req *pb.FunctionWithEmptyParamRequest) (ret *pb.FunctionWithEmptyParamResponse, err error) {
	ret, err = setFunctionWithEmptyParam(req)
	if err != nil {
		return ret, err
	}
	return ret, err
}

func ReflectStructField(Iface interface{}, FieldName string) error {
	ValueIface := reflect.ValueOf(Iface)

	// Check if the passed interface is a pointer
	if ValueIface.Type().Kind() != reflect.Ptr {
		// Create a new type of Iface's Type, so we have a pointer to work with
		ValueIface = reflect.New(reflect.TypeOf(Iface))
	}

	// 'dereference' with Elem() and get the field by name
	Field := ValueIface.Elem().FieldByName(FieldName)
	if !Field.IsValid() {
		return fmt.Errorf("Interface `%s` does not have the field `%s`", ValueIface.Type(), FieldName)
	}
	return nil
}
func (s *GRPCServer) GetFlagOfType(ctx context.Context, req *pb.FlagOfTypeRequest) (*pb.FlagOfTypeResponse, error) {
	if req.GetType() <= 0 || req.GetType() > 6 {
		return &pb.FlagOfTypeResponse{}, errors.New("Передан не верный тип.")
	}
	r, err := getFlagOfTypeList(db, int(req.GetType()))
	if err != nil {
		return &pb.FlagOfTypeResponse{}, err
	}
	return r, err
}

type functionBigByte struct {
	Byte [1]byte
}

func getFuctionCode(bd []byte) (code_func uint16, resource_inx uint16, error_type bool, err error) {
	error_type = false
	code_func = binary.LittleEndian.Uint16(bd[0:2])
	index := code_func >> 13
	resource_inx = 1
	if index == 4 && len(bd) == 4 {
		error_type = true
	}
	if !error_type { //если не ошибка, то номер ресурса
		resource_inx = index + 1
	}
	code_func = code_func ^ (index << 13)
	return code_func, resource_inx, error_type, err
}

func getPacketError(dst []byte, vr string) (b bool, error_const string, err error) {
	b = false
	var ep ErrorByteRX
	buf := bytes.NewReader(dst)
	err = binary.Read(buf, binary.BigEndian, &ep)
	if err != nil {
		log.Println("binary.Read failed:", err)
		return
	}
	byte_packet := binary.LittleEndian.Uint16(ep.ErrorType[:])

	var epack map[string][]byte

	if _, ok := error_packet[vr]; ok {
		epack = error_packet[vr]
	} else if vr == "1.6" || vr == "1.7" || vr == "1.8" {
		epack = error_packet["1.8"]
	} else {
		epack = error_packet["1.9"]
	}

	for error_name, error_byte := range epack {
		error_packet_uint := binary.BigEndian.Uint16(error_byte)
		if byte_packet == error_packet_uint {
			b = true
			return b, error_name, nil
		}
	}
	return true, "Ошибка не определена", nil
}

type ErrorByteRX struct {
	CodeFunction [2]byte
	ErrorType    [2]byte
}

func byteToBits(b [1]byte) (bits []int) {
	x := b[0]
	for i := uint(0); i < 8; i++ {
		bits = append(bits, int(x&(1<<i)>>i))
	}
	return
}

func str2Byte(command string) (ret []byte, err error) {
	src := []byte(command)
	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err = hex.Decode(dst, src)
	if err != nil {
		log.Println(err)
		return
	}
	return dst[:], err
}
func packetDecode(code int, data []byte) (ret []*pb.Data, err error) {
	//6e020f170d04148e71ad3ebea4a7420e1fa642c5b14142408831421a4c81406f12833a4260e53b295c0f3d5c880100dcdb0600
	var time_ float32
	var requestDate2 TIME_TYPE
	var rvalue_id, resource_id int32
	if _, ok := conf[code]; !ok {
		return ret, errors.New(fmt.Sprintf("Отсутствует описание пакета для функции %v", code))
	}
	conf_ := conf[code]
	for _, item := range conf_.Scans {
		var rec pb.Data
		var inter_value interface{}
		if !item.RvalueId.Valid {
			continue
		}
		rvalue_id = int32(item.RvalueId.Int64)
		resource_id = int32(item.ResourceId.Int64)
		index := item.Index
		byte_for := index - 1
		byte_len := item.Len
		if err != nil {
			log.Println(err)
			return nil, err
		}
		if byte_for+byte_len > len(data) {
			log.Println("Байт меньше чем в конфигурации")
			return nil, errors.New("Байт меньше чем в конфигурации")
		}
		bd_ := data[byte_for : byte_for+byte_len]
		var bd = make([]byte, item.Len)
		copy(bd[:], bd_)
		if item.MaskType == "t" {
			time_ = 60
			if contains(conf_.ConfModel, "213") {
				time_ = 600
			} /* else if contains(conf_.ConfModel, "308") {
				time_ = 360
			}*/
			inter_value = float32(binary.LittleEndian.Uint32(bd)) / time_
		} else if item.TypeData == data_types["int16"] {
			var y int16
			if item.LittleEndian {
				_ = binary.Read(bytes.NewReader(bd), binary.LittleEndian, &y)
			} else {
				_ = binary.Read(bytes.NewReader(bd), binary.BigEndian, &y)
			}
			inter_value = y
		} else if item.TypeData == data_types["int32"] {
			var y int32
			if item.LittleEndian {
				_ = binary.Read(bytes.NewReader(bd), binary.LittleEndian, &y)
			} else {
				_ = binary.Read(bytes.NewReader(bd), binary.BigEndian, &y)
			}
			inter_value = y
		} else if item.TypeData == data_types["int64"] {
			var y int64
			if item.LittleEndian {
				_ = binary.Read(bytes.NewReader(bd), binary.LittleEndian, &y)
			} else {
				_ = binary.Read(bytes.NewReader(bd), binary.BigEndian, &y)
			}
			inter_value = y
		} else if item.TypeData == data_types["float32"] {
			if item.LittleEndian {
				inter_value = math.Float32frombits(binary.LittleEndian.Uint32(bd))
			} else {
				inter_value = math.Float32frombits(binary.BigEndian.Uint32(bd))
			}
		} else if item.TypeData == data_types["byte2bit"] {
			if item.MaskType == "arch_types" {
				inter_value = getFlag(bd[:1])
			} else if item.MaskType == "con_status_flags" {

			} else {
				inter_value = getFlag(bd[:1])
			}
		} else if item.TypeData == data_types["int8"] {
			var y int8
			if item.LittleEndian {
				_ = binary.Read(bytes.NewReader(bd), binary.LittleEndian, &y)
			} else {
				_ = binary.Read(bytes.NewReader(bd), binary.BigEndian, &y)
			}
			inter_value = y
		} else if item.TypeData == data_types["uint16"] {
			if item.LittleEndian {
				inter_value = binary.LittleEndian.Uint16(bd)
			} else {
				inter_value = binary.BigEndian.Uint16(bd)
			}
		} else if item.TypeData == data_types["uint32"] {
			if item.LittleEndian {
				inter_value = binary.LittleEndian.Uint32(bd)
			} else {
				inter_value = binary.BigEndian.Uint32(bd)
			}
		} else if item.TypeData == data_types["uint64"] {
			if item.LittleEndian {
				inter_value = binary.LittleEndian.Uint64(bd)
			} else {
				inter_value = binary.BigEndian.Uint64(bd)
			}
		} else if item.TypeData == data_types["uint8"] {
			value, err := ByteToUint8(bd)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			inter_value = value
		} else if item.TypeData == data_types["Uint24"] {
			inter_value = int24ToInt32(reverseBytes(bd))
		} else if item.TypeData == data_types["protoVR"] {
			var vf float32
			if s.Contains(conf_.ConfVersion, "1.8") {
				vr, _ := strconv.Atoi(strconv.FormatInt(int64(bd[0]), 16))
				vf = float32(vr) / 10
			} else {
				value, err := ByteToUint8(bd)
				if err != nil {
					log.Println(err)
					return nil, err
				}
				if value < 26 {
					vf = float32(value) / 10
				} else {
					vf = (float32(value) - 6) / 10
				}
			}
			inter_value = fmt.Sprintf("%v", vf)
		} else if item.TypeData == data_types["archs"] {
			inter_value = getArch(bd)
		} else if item.TypeData == data_types["error"] {
			inter_value = getCheckErrorNew(bd, conf_)
		} else if item.TypeData == data_types["NodeSerialNum1.8"] {
			snImp, _ := strconv.ParseUint(fmt.Sprintf("0x%x", reverseBytes(bd)), 0, 64)
			inter_value = fmt.Sprintf("%v", snImp)
		} else if item.TypeData == data_types["TimeZone"] {
			var y int8
			_ = binary.Read(bytes.NewReader(bd), binary.BigEndian, &y)
			inter_value = float32(y) * 15 / 60
		} else if item.TypeData == data_types["EventMaskFlags"] {
			inter_value = getEventFlag(bd)
		} else if item.TypeData == data_types["ConStatusFlags"] {
			inter_value = getConFlag(bd)
		} else if item.TypeData == data_types["CFlags"] {
			inter_value = getLoraFlag(bd)
		} else if item.TypeData == data_types["HDMY"] {
			var time_arch TIME_HDMY_TYPE
			if s.Contains(conf_.ConfVersion, "1.9") || s.Contains(conf_.ConfVersion, "2.0") {
				lw_time_ext := int64(binary.LittleEndian.Uint32(bd))
				sec_time := Second2LwTimeExtREAD(lw_time_ext)
				requestDate2.Hour = sec_time.Hour()
				requestDate2.Day = sec_time.Day()
				requestDate2.Month = int(sec_time.Month())
				requestDate2.Year = sec_time.Year()
			} else {
				bd := bytes.NewBuffer(bd)
				err := binary.Read(bd, binary.LittleEndian, &time_arch)
				if err != nil {
					log.Printf("Error: %s\n", err)
					return ret, err
				}
				requestDate2.Hour = int(time_arch.Hour)
				requestDate2.Day = int(time_arch.Day)
				requestDate2.Month = int(time_arch.Month)
				requestDate2.Year = int(2000 + int(time_arch.Year))
			}
			inter_value = requestDate2.struct_to_time()
		} else if item.TypeData == data_types["DMY"] {
			var time_arch TIME_DMY_TYPE
			if s.Contains(conf_.ConfVersion, "1.9") || s.Contains(conf_.ConfVersion, "2.0") {
				lw_time_ext := int64(binary.LittleEndian.Uint32(bd))
				sec_time := Second2LwTimeExtREAD(lw_time_ext)
				requestDate2.Day = sec_time.Day()
				requestDate2.Month = int(sec_time.Month())
				requestDate2.Year = int(sec_time.Year())
			} else {
				bd := bytes.NewBuffer(bd)
				err := binary.Read(bd, binary.LittleEndian, &time_arch)
				if err != nil {
					log.Printf("Error: %s\n", err)
					return ret, err
				}
				requestDate2.Day = int(time_arch.Day)
				requestDate2.Month = int(time_arch.Month)
				requestDate2.Year = 2000 + int(time_arch.Year)
			}
			inter_value = requestDate2.struct_to_time()
		} else if item.TypeData == data_types["MY"] {
			var time_arch TIME_MY_TYPE
			if s.Contains(conf_.ConfVersion, "1.9") || s.Contains(conf_.ConfVersion, "2.0") {
				lw_time_ext := int64(binary.LittleEndian.Uint32(bd))
				sec_time := Second2LwTimeExtREAD(lw_time_ext)
				requestDate2.Month = int(sec_time.Month())
				requestDate2.Year = sec_time.Year()
			} else {
				bd := bytes.NewBuffer(bd)
				err := binary.Read(bd, binary.LittleEndian, &time_arch)
				if err != nil {
					log.Printf("Error: %s\n", err)
					return ret, err
				}
				requestDate2.Month = int(time_arch.Month)
				requestDate2.Year = 2000 + int(time_arch.Year)
			}
			inter_value = requestDate2.struct_to_time()
		} else if item.TypeData == data_types["SMHDMY"] {
			var time_arch TIME_SMHDMY_TYPE
			if s.Contains(conf_.ConfVersion, "1.9") || s.Contains(conf_.ConfVersion, "2.0") {
				lw_time_ext := int64(binary.LittleEndian.Uint32(bd))
				sec_time := Second2LwTimeExtREAD(lw_time_ext)
				requestDate2.Sec = sec_time.Second()
				requestDate2.Min = sec_time.Minute()
				requestDate2.Hour = sec_time.Hour()
				requestDate2.Day = sec_time.Day()
				requestDate2.Month = int(sec_time.Month())
				requestDate2.Year = sec_time.Year()
			} else {
				bd := bytes.NewBuffer(bd)
				err := binary.Read(bd, binary.LittleEndian, &time_arch)
				if err != nil {
					log.Printf("Error: %s\n", err)
					return ret, err
				}
				requestDate2.Sec = int(time_arch.Sec)
				requestDate2.Min = int(time_arch.Min)
				requestDate2.Hour = int(time_arch.Hour)
				requestDate2.Day = int(time_arch.Day)
				requestDate2.Month = int(time_arch.Month)
				requestDate2.Year = 2000 + int(time_arch.Year)
			}
			inter_value = requestDate2.struct_to_time()
		}

		if item.Multiplier.Valid && item.Multiplier.Float64 != 0 && item.Multiplier.Float64 != 1 {
			var k float64 = 100000000
			switch i := inter_value.(type) {
			case int32:
				inter_value = float64(int64(i)*int64(item.Multiplier.Float64*k)) / k
				break
			case uint32:
				inter_value = float64(uint64(i)*uint64(item.Multiplier.Float64*k)) / k
				break
			case float32:
				inter_value = float64(i) * item.Multiplier.Float64
				break
			default:
				return ret, errors.New("Multiplier: unknown value is of incompatible type")
			}
		}
		rec = pb.Data{RvalueId: rvalue_id, ResourceId: resource_id, Alias: item.Alias, Value: ToValue(inter_value)}
		ret = append(ret, &rec)
	}
	return ret, err
}
func int24ToInt32(bs []byte) uint32 {
	return uint32(bs[2]) | uint32(bs[1])<<8 | uint32(bs[0])<<16
}
func contains(arr []string, str string) bool {
	if len(arr) == 0 {
		return false
	}
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
func getCheckErrorNew(error_struct []byte, conf_ ConfStruct) (ret []int /*FlagStruct*/) {
	if len(conf_.ConfModel) == 0 {
		return
	}
	if _, ok := error_archive[conf_.ConfModel[0]]; !ok {
		return
	}
	var ee []int
	for _, item := range error_struct {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item := range ee {
		if item == 1 {
			if _, ok := error_archive[conf_.ConfModel[0]][index]; !ok {
				continue
			}
			ret = append(ret, error_archive[conf_.ConfModel[0]][index].Id)
		}
	}
	return ret
}
func reverseBytes(by []byte) []byte {
	if len(by) == 0 {
		return by
	}
	return append(reverseBytes(by[1:]), by[0])
}
func getArch(arch_byte []byte) (ret []int) {
	var (
		ee []int
	)
	for _, item := range arch_byte {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item2 := range ee {
		if item2 == 1 {
			ret = append(ret, arch_byte_const[index])
		}
	}
	return ret
}
func getFlag(val []byte) (ret []int) {
	sb, err := str2Byte(fmt.Sprintf("%X", val[:1]))
	if err != nil {
		log.Println(err)
		return
	}
	var fbb [1]byte
	buf := bytes.NewReader(sb[:])
	err = binary.Read(buf, binary.BigEndian, &fbb)
	if err != nil {
		log.Println(err)
		return
	}
	return byteToBits(fbb)
}
func getFlagStatusNByte(status_byte []byte) (ret []int) {
	var (
		ee []int
	)
	for _, item := range status_byte {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item2 := range ee {
		if item2 == 1 {
			ret = append(ret, index)
		}
	}
	return ret
}
func getEventFlag(error_struct []byte) (ret []int) {
	var ee []int
	for _, item := range error_struct {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item := range ee {
		if item == 1 {
			if _, ok := event_flag[index]; !ok {
				continue
			}
			ret = append(ret, event_flag[index].Id)
		}
	}
	return ret
}
func getConFlag(con_byte []byte) (ret []int) {
	var ee []int
	for _, item := range con_byte {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item := range ee {
		if item == 1 {
			if _, ok := con_flag[index]; !ok {
				continue
			}
			ret = append(ret, con_flag[index].Id)
		}
	}
	return ret
}
func getLoraFlag(lora_byte []byte) (ret []int) {
	var ee []int
	for _, item := range lora_byte {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item := range ee {
		if item == 1 {
			if _, ok := lora_flag[index]; !ok {
				continue
			}
			ret = append(ret, lora_flag[index].Id)
		}
	}
	return ret
}
func getFlagStatusNByteTemp(status_byte []byte) (ret int) {
	var (
		ee []int
	)
	for _, item := range reverseBytes(status_byte) {
		ee = append(ee, byteToBits([1]byte{item})...)
	}
	for index, item2 := range ee {
		if item2 == 1 {
			ret = index
			break
		}
	}
	return ret
}
func ByteToUint8(b []byte) (val uint8, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.LittleEndian, &val)
	if err != nil {
		log.Println("binary.Read failed:", err)
		return
	}
	return
}
func Second2LwTimeExtREAD(sec int64) (date time.Time) {
	date = time.Date(2000, 03, 01, 0, 0, 0, 0, time.UTC)
	date = date.Add(time.Second * time.Duration(sec))
	return
}

func (t *TIME_TYPE) struct_to_time() (datetime string) {
	var tt time.Time
	tt = time.Date(t.Year, time.Month(t.Month), t.Day, t.Hour, t.Min, t.Sec, 0, time.UTC)
	return tt.Format("2006-01-02 15:04:05")
}
func getByte(byte_str string) (ret []byte, err error) {
	s := strings.Split(byte_str, ",")
	for _, value := range s {
		value = strings.Replace(value, " ", "", -1)
		u64, err := strconv.ParseUint(value, 0, 16)
		if err != nil {
			log.Println(err)
			return ret, err
		}
		ret = append(ret, uint8(u64))
	}
	return
}
