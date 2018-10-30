package shared

import "fmt"
import "encoding/json"
import "encoding/binary"
import "os"
import "os/exec"
import "strings"

func ConvJsonToDataStruct(jsonData []byte) *DataObj {
	fmt.Printf("\n Converting json data: % x", jsonData)

	dataObj := new(DataObj)

	// FIXED
	// extract type field and add to struct
	typeField := jsonData[0:2]
	dataObj.Type = uint64(binary.BigEndian.Uint16(typeField))

	jsonB :=  jsonData[2:]
	err := json.Unmarshal(jsonB, dataObj)
	if err != nil {
		fmt.Printf("Cannot Unmarshal %s\n", err)
		os.Exit(1)
	}

	return dataObj
}

func ConvDataStructToJson(data *DataObj) []byte {
	fmt.Println("\nConverting datastruct ", *data)

	var resB []byte

	// construct type field
	typeB := make([]byte, 2)
	binary.BigEndian.PutUint16(typeB[0:2], uint16(data.Type))

	// ignore Type
	data.Type = 0

	jsonB, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Cannot Marshal %s\n", err)
		os.Exit(1)
	}

	resB = append(resB, typeB...)
	resB = append(resB, jsonB...)
	return resB
}

func ConstructId() string {
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Printf("Cannot get hostname %s\n", err)
		os.Exit(1)
	}

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Printf("Cannot construct uuid %s\n", err)
		os.Exit(1)
	}

	id := []string{}
	id = append(id, hostName)
	id = append(id, strings.TrimSuffix(string(uuid[:]), "\n"))
	return strings.Join(id, "=")
}