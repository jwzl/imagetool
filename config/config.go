package config

import(
	"io"
	"os"
	"fmt"
	"bufio"
	"strconv"
	"strings"
)

func PackageFileLoad(path string) map[string]string{
	config := make(map[string]string)

	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("open %s with err:%v \r\n", path, err)
		return nil
	}
	defer f.Close()
	
	r := bufio.NewReader(f)
	for{
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break;
			}
			
			return config
		}
		
		fields := strings.Fields(string(line))
		if len(fields) < 1 {
			continue
		}else if len(fields) == 1 {
			key := strings.SplitN(fields[0], ":", 2)[0]
			index := strings.Index(key, "#")
			if index != 0 {
				key = key[0:index]
				config[key] = ""
			}
			continue
		}
		
		key := strings.SplitN(fields[0], ":", 2)[0]
		value	:= fields[1]
		
		//ignore the mask the "#"
		index := strings.Index(key, "#")
		if index >= 0 {
			if index != 0 {
				key = key[0:index]
				config[key] = ""
			}
			continue
		}
		
		index = strings.Index(value, "#")
		if index > 0 {
			value = value[0:index]
		}else if index == 0 {
			value = ""
		}
		
		config[key] = value
	}
	
	return config
}

type DiskParameter struct {
	FirmwareVersion		string
	MachineModel		string
	MachineID			string
	Manufacturer		string
	Magic				string
	ATAG				string
	Machine				string
	CheckMask			string
	PowerHold			string
	Type				string
	CommandLine			string
}

/*
* PartitionInfo:
* Partition Infomation.
*/
type PartitionInfo struct{
	Name		string
	Offset		uint64
	Length		uint64
}

func ParameterLoad(path string) *DiskParameter {
	dp := &DiskParameter{}
	
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("open %s with err:%v \r\n", path, err)
		return nil
	}
	defer f.Close()
	
	r := bufio.NewReader(f)
	for{
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break;
			}
			return dp
		}

		sLine := strings.TrimSpace(string(line))
		sLine = strings.ReplaceAll(sLine, " ", "")
		level := strings.SplitN(sLine, ":", 2)

		if len(level) <= 1 {
			continue
		}

		key := level[0]
		value := level[1]

		index := strings.Index(key, "#")
		if index >= 0 {
			continue
		}

		switch {
		case strings.Contains(key, "FIRMWARE_VER"):
			dp.FirmwareVersion = value
		case strings.Contains(key, "MACHINE_MODEL"):
			dp.MachineModel = value
		case strings.Contains(key, "MACHINE_ID"):
			dp.MachineID = value
		case strings.Contains(key, "MANUFACTURER"):
			dp.Manufacturer = value
		case strings.Contains(key, "MAGIC"):
			dp.Magic = value
		case strings.Contains(key, "ATAG"):
			dp.ATAG = value
		case strings.Contains(key, "MACHINE"):
			dp.Machine = value
		case strings.Contains(key, "CHECK_MASK"):
			dp.CheckMask = value
		case strings.Contains(key, "PWR_HLD"):
			dp.PowerHold = value
		case strings.Contains(key, "TYPE"):
			dp.Type = value
		case strings.Contains(key, "CMDLINE"):
			dp.CommandLine = value
		}
	}

	return dp	
}

/*
* GetDiskPartition:
* get the disk info.
*/
func (dp *DiskParameter) GetDiskPartition() []*PartitionInfo {
	commandLine := dp.CommandLine
	
	levels := strings.SplitN(commandLine, "=", 2)
	if len(levels) != 2 {
		return nil
	}

	mtdparts := levels[1]
	values := strings.SplitN(mtdparts, ":", 2)
	if len(values) != 2 {
		return nil
	}

	partitionInfos := make([]*PartitionInfo, 0)
	partsInfo := values[1]
	
	parts := strings.Split(partsInfo, ",")
	for _, part := range parts {
		v := strings.SplitN(part, "@", 2)
		if len(v) != 2 {
			continue
		}

		length, _ := strconv.ParseUint(v[0], 0, 64)

		tmp := v[1]
		if strings.Index(tmp, "(") < 0 || strings.Index(tmp, ")") < 0 ||
			strings.Index(tmp, "(")  == 0 ||
			(strings.Index(tmp, "(") +1 >= len(tmp)) ||
			(strings.Index(tmp, "(") +1 >= strings.Index(tmp, ")")){
			continue
		}

		o := tmp[0:strings.Index(tmp, "(")]
		offset, err := strconv.ParseUint(o, 0, 64)
		if err != nil {
			continue
		}

		name := tmp[strings.Index(tmp, "(")+1 : strings.Index(tmp, ")")]
		name = strings.SplitN(name, ":", 2)[0]

		//add partition.
		partInfo := &PartitionInfo{
			Name:	name,
			Offset:	offset,
			Length:	length,
		}
		partitionInfos = append(partitionInfos, partInfo)
	}

	return partitionInfos
}

// GetDiskPartitionInfos
func GetDiskPartitionInfos(filePath string) []*PartitionInfo {
	diskInfo := ParameterLoad(filePath)
	if diskInfo == nil {
		return nil
	}

	return diskInfo.GetDiskPartition() 
}

func CheckPartitionIsExist(parts []*PartitionInfo, name string) bool {
	for _, part := range parts {
		if part == nil {
			continue
		}

		if part.Name == name {
			return true
		}
	}

	return false
}
