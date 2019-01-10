package shared

import "fmt"
import "encoding/json"
import "encoding/binary"
import "os"
import "os/exec"
import "strings"
import "bytes"
import "time"
import "runtime"
import "path/filepath"

const DATE_FMT = "2006-01-02 15:04:05.000000000"

func ConvJsonToDataStruct(jsonData []byte) *DataObj {
	// debug fmt.Printf("\n Converting json data: % x", jsonData)

	dataObj := new(DataObj)

	// FIXED
	// extract type field and add to struct
	typeField := jsonData[0:2]
	dataObj.Type = uint64(binary.BigEndian.Uint16(typeField))

	jsonB := jsonData[2:]
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
		fmt.Printf("\nCannot Marshal %s\n", err)
		os.Exit(1)
	}

	resB = append(resB, typeB...)
	resB = append(resB, jsonB...)
	return resB
}

// RFC: the measurement listening port must not configured here
func ConstructConfiguration(configDir string) *ConfigurationObj {
	ConfObj := new(ConfigurationObj)

	absPath, err := filepath.Abs(configDir)
	if err != nil {
		fmt.Printf("\nError while calc abs path: %s", err)
		os.Exit(1)
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) == true {
		fmt.Printf("\nConfig not present! Use default one!")

		ConfObj.Worker = "1"
		ConfObj.Port = "7000"
		ConfObj.Listen_addr = "127.0.0.1"
		ConfObj.Call_size = "64768"

	} else {
		fmt.Printf("\nConfig present! Decode it!")

		fPtr, err := os.Open(absPath)
		if err != nil {
			fmt.Printf("\nError while opening measurement Config: %s", err)
			os.Exit(1)
		}

		defer fPtr.Close()

		decoderPtr := json.NewDecoder(fPtr)
		err = decoderPtr.Decode(ConfObj)
		if err != nil {
			fmt.Printf("\nError while decoding: %s", err)
			os.Exit(1)
		}
	}
	return ConfObj
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

func ConvCurrDateToStr() string {
	dateStr := time.Now().Format(DATE_FMT)
	return dateStr
}

func ConvIntSliceToStr(slice []int) string {
	str := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(slice)), ","), "[]")
	return str
}

func ConvStrDateToNatDate(date string) time.Time {
	time, err := time.Parse(DATE_FMT, date)
	if err != nil {
		fmt.Printf("Cannot parse str to time % s \n", err)
		os.Exit(1)
	}

	return time
}

func DetectOs() string {
	os := runtime.GOOS
	return os
}

func DetectArch() string {
	arch := runtime.GOARCH
	return arch
}

func ConvMapToStr(m map[string]string) string {
	buf := new(bytes.Buffer)

	for key, value := range m {
		_, err := fmt.Fprintf(buf, "%s:%s ", key, value)
		if err != nil {
			fmt.Printf("Cannot conv map to str: %s \n", err)
			os.Exit(1)
		}
	}
	return buf.String()
}
