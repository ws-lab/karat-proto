package karatproto

import (
	"database/sql"
)

type TIME_TYPE struct {
	Sec   int
	Min   int
	Hour  int
	Day   int
	Week  int
	Month int
	Year  int
}
type TIME_HDMY_TYPE struct {
	Hour  uint8
	Day   uint8
	Month uint8
	Year  uint8
}

type TIME_DMY_TYPE struct {
	Day   uint8
	Month uint8
	Year  uint8
}

type TIME_MY_TYPE struct {
	Month uint8
	Year  uint8
}
type TIME_SMHDMY_TYPE struct {
	Sec   uint8
	Min   uint8
	Hour  uint8
	Day   uint8
	Month uint8
	Year  uint8
}

type ConfStruct struct {
	Scans       []Scan
	ConfVersion string
	ConfModel   []string
	PacketType  int
}

type FlagStruct struct {
	Id   int
	Note string
	Byte string
}

type RetPacket struct {
	RvariableId int           `json:"rvariable_id"`
	ResourceId  sql.NullInt64 `json:"resource_id"`
	RvalueId    sql.NullInt64 `json:"rvalue_id"`
	IsSensor    bool          `json:"is_sensor"`
	Alias       string        `json:"alias"`
	Value       interface{}   `json:"value"`
}
type Return struct {
	RetPackets []RetPacket
	Code       int    `db:"code"`
	VR         string `db:"vr"`
	Models     string `db:"dmodels"`
	Alias      string `db:"alias"`
}
type Scan struct {
	//DconfExtId int    `json:"dconf_ext_id"`
	Code         int             `db:"code"`
	Index        int             `db:"index"`
	ResourceId   sql.NullInt64   `db:"resource_id"`
	RvalueId     sql.NullInt64   `db:"rvalue_id"`
	MaskType     string          `db:"mask_type"`
	Len          int             `db:"len"`
	TypeData     string          `db:"type_data"`
	LittleEndian bool            `db:"little_endian"`
	IsSensor     bool            `db:"is_sensor"`
	VR           string          `db:"vr"`
	Models       string          `db:"dmodels"`
	Alias        string          `db:"alias"`
	PacketType   int             `db:"packet_type"`
	Multiplier   sql.NullFloat64 `db:"multiplier"`
}
type Packet struct {
	Id         int32  `db:"id"`
	Vr         string `db:"vr"`
	Models     string `db:"dmodels"`
	PacketType int32  `db:"packet_type"`
}
type PacketConf struct {
	Id           int32         `db:"id"`
	Code         int32         `db:"code"`
	Index        int32         `db:"index"`
	ResourceId   int32         `db:"resource_id"`
	RvariableId  int32         `db:"rvariable_id"`
	RvalueId     sql.NullInt64 `db:"rvalue_id"`
	Len          int32         `db:"len"`
	TypeData     int32         `db:"type_data"`
	LittleEndian bool          `db:"little_endian"`
	IsSensor     bool          `db:"is_sensor"`
	UnitId       sql.NullInt64 `db:"unit_id"`
}
type Rvalue struct {
	Id          int32         `db:"id"`
	Name        string        `db:"name"`
	ResourceId  int32         `db:"resource_id"`
	RvariableId int32         `db:"rvariable_id"`
	UnitId      sql.NullInt64 `db:"unit_id"`
}
type Flags struct {
	Id      int32  `db:"id"`
	Note    string `db:"note"`
	Byte    string `db:"byte"`
	Nbit    int64  `db:"nbit"`
	Devices string `db:"device"`
}

var data_types = map[string]string{
	"int16":    "1",
	"int32":    "2",
	"int64":    "3",
	"float32":  "4",
	"byte2bit": "5",
	"int8":     "6",
	"uint16":   "8",
	"uint32":   "9",
	"uint64":   "10",
	"uint8":    "11",
	"SMHDMY":   "12", // karat lora, 6 байт
	"DMY":      "13", // PZIP lora, 3 байтa
	"MY":       "14", // PZIP lora, 2 байтa
	"error":    "15", //4 байта
	//"error_926":      "16", //1 байт
	"HDMY":           "17", // karat lora, 4 байта
	"protoVR":        "18", // karat lora, 1 байт
	"archs":          "19", //karat lora, 1 байт типы архивов
	"ConStatusFlags": "20", //karat lora, 2 байта флаги статуса связи, причина передачи пакета,0x0100
	"EventMaskFlags": "21", //karat lora, 2 байта Маска флагов событий выхода на связь,0x0101
	//"SMHDMY2.0":      "22", // karat lora, 4 байта
	"TimeZone":         "23", // karat lora, 2 байта
	"NodeSerialNum1.8": "24", // karat lora, 6 байт
	"CFlags":           "25", // karat lora, 1 байт флаги лоры
	"Uint24":           "26", // karat lora, 3 байта
	"LogCode":          "27", // karat lora, 4 байта
}

var error_926 = []int{
	0: 100,
	1: 101,
	2: 102,
	3: 103,
	4: 104,
	5: 105,
	6: 106,
	7: 107,
}

var arch_byte_const = map[int]int{
	0: 5,
	1: 6,
	2: 7,
	3: 8,
	4: 9,
	5: 14,
	6: 15,
	7: 16,
}

var karat_type2system_type = map[int]int{
	5:  1,
	6:  2,
	7:  3,
	8:  5,
	9:  4,
	14: 6,
	15: 7,
}

//0x0101
type SettingsPacketByteTX20 struct {
	CodeFunction   [2]byte
	EventMaskFlags [2]byte
	DeltaDevTime   [4]byte
	TimeZone       int8
	OffsetTime     [2]byte
	TxPeriod       [2]byte
	ArchType       [1]byte
	MainMsgCnt     uint8
	MMsg1          [2]byte
	MMsg2          [2]byte
	MMsg3          [2]byte
	MMsg4          [2]byte
	MMsg5          [2]byte
	MMsg6          [2]byte
}

//0x0015
type SettingsPacketByteTX18 struct {
	CodeFunction [2]byte
	LWTime       LWTime
	DeltaDevTime [2]byte
	OffsetTime   [4]byte
	TxPeriod     [4]byte
	ArchType     [1]byte
	MainMsgCnt   uint8
	MMsg1        [2]byte
	MMsg2        [2]byte
	MMsg3        [2]byte
	MMsg4        [2]byte
	MMsg5        [2]byte
	MMsg6        [2]byte
	TimeZone     uint8
	BegPeriod    uint8
}
type LWTime struct {
	Sec   uint8
	Min   uint8
	Hour  uint8
	Day   uint8
	Month uint8
	Year  uint8
}

//0x0016
type TimeCorrectByteTx struct {
	CodeFunction [2]byte
	DeltaDevTime [2]byte
}

type setArchivePacketByteTx18 struct {
	CodeFunction [2]byte
	ArchType     uint8
	Hour         uint8
	Day          uint8
	Month        uint8
	Year         uint8
}
type setArchivePacketByteTx20 struct {
	CodeFunction [2]byte
	ArchType     uint8
	ArchTime     [4]byte
}
type setLoraWanParamsByteTx20 struct {
	CodeFunction [2]byte
	JoinParams   uint8
	DN2Freq      [3]byte
	Port         uint8
	TxAtts       uint8
	ADRparam     uint8
	CFlags       [1]byte
}
