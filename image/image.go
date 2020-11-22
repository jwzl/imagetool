package image

import (
	"io"
	"os"
	"fmt"
	"time"
	"errors"
	"strings"
	"crypto/md5"
	"encoding/hex"
	"encoding/binary"

	"github.com/jwzl/imagetool/config"
)

const (
	RKNONE_DEVICE=0
    RK27_DEVICE=0x10
    RKCAYMAN_DEVICE=0x11
    RK28_DEVICE=0x20
    RK281X_DEVICE=0x21
    RKPANDA_DEVICE=0x22
    RKNANO_DEVICE=0x30
    RKSMART_DEVICE=0x31
    RKCROWN_DEVICE=0x40
    RK29_DEVICE=0x50
    RK292X_DEVICE=0x51
    RK30_DEVICE=0x60
    RK30B_DEVICE=0x61
    RK31_DEVICE=0x70
    RK32_DEVICE=0x80

	IMAGE_VERSION = 0x8010000
)


type RKTime struct{
	Year		uint16
	Mouth		byte
	Day			byte
	Hour		byte
	Minute		byte
	Second		byte
}

//make rk time.
func MakeRKTime() RKTime {
	t := time.Now()
	
	return RKTime{
		Year: uint16(t.Year()),
		Mouth:	byte(t.Month()),
		Day:	byte(t.Day()),
		Hour:	byte(t.Hour()),
		Minute:	byte(t.Minute()),
		Second: byte(t.Second()),
	}
}

type RKImageHeader struct{
	Tag				uint32		//tag, fixed as 0x57 0x46 0x4B 0x52
	Size			uint16		//size of the struct
	Version			uint32		//image version
	ToolVersion		uint32		//version of package tool
	ReleaseTime		RKTime		//the time of do package
	SupportChip		uint32		//support chip
	BootOffset		uint32		//the offset of the bootloader in this image
	BootLength		uint32		//the length of the bootloader.
	FirmwareOffset	uint32		//the offset of the firmware in this image
	FirmwareLength	uint32		//the length of the firmware.
	Reserved		[61]byte	//reserved for the firmware setting
}
func NewRKImageHeader(version, toolVersion, supportChip uint32) *RKImageHeader {
	
	header := &RKImageHeader{
		Tag:	0x57464B52,
		Version: version,
		ToolVersion:	toolVersion,
		ReleaseTime:	MakeRKTime(),
		SupportChip:	supportChip,
	}
	
	return header
}

func (ri *RKImageHeader) SetBootLoaderLocation(offset, length uint32){
	ri.BootOffset = offset
	ri.BootLength = length
}

func (ri *RKImageHeader) SetFirmwareLocation(offset, length uint32){
	ri.FirmwareOffset = offset
	ri.FirmwareLength = length
}

/*
* WriteRKImageHeader:
* write the rockchip image header.
*/
func WriteRKImageHeader(w io.Writer, header *RKImageHeader) error {
	if 	header == nil {
		return errors.New("header is nil")
	}
	
	//write to writer.
	err := binary.Write(w, binary.LittleEndian, *header)
	if err != nil {
		return err
	}
	
	return nil	
}

type ImageItem	struct{
	ImageName		[32]byte		//image name
	FilePath		[64]byte		//image file path 
	ImageOffset		uint32			//image offset
	FlashOffset		int32			//flash offset
	UseSpace		uint32			
	Length			uint32			//image Length
}

func NewImageItem(name, filePath string, offset, lenght, useSpace uint32, flashOffset int32) *ImageItem {
	item := &ImageItem{
		ImageOffset: offset,
		FlashOffset: flashOffset,
		UseSpace: useSpace,
		Length: lenght,
	}

	for i := int(0); i < len(name); i++ {
		if  i < len(item.ImageName) {
			item.ImageName[i] = byte(name[i])
		}
	}

	for i := int(0); i < len(filePath); i++ {
		if  i < len(item.FilePath) {
			item.FilePath[i] = byte(filePath[i])
		}
	}

	return item
}

/*
* ImageHeader:
* firmware image header.
*/
type ImageHeader struct{
	Tag				uint32		//0x46414B52
	Size			uint32
	MachineModel	[64]byte
	Manufacturer	[60]byte
	Version			uint32
	ItemCount		int32
	ImageItems		[16]ImageItem
}

func NewImageHeader(machineModel, manufacturer string) *ImageHeader {
	header := &ImageHeader{
		Tag: 0x46414B52,	
		Version: IMAGE_VERSION,
	}

	for i := int(0); i < len(machineModel); i++ {
		if  i < len(header.MachineModel) {
			header.MachineModel[i] = byte(machineModel[i])
		}
	}

	for i := int(0); i < len(manufacturer); i++ {
		if  i < len(header.Manufacturer) {
			header.Manufacturer[i] = byte(manufacturer[i])
		}
	}

	return header
}


func WriteImageHeader(w io.Writer, header *ImageHeader) error {
	if 	header == nil {
		return errors.New("header is nil")
	}
	
	//write to writer.
	err := binary.Write(w, binary.LittleEndian, *header)
	if err != nil {
		return err
	}
	
	return nil	
}

func CaculateFileSize(fileName string) int64{
	info, err := os.Stat(fileName)
	if err != nil {
		return 0
	}

	if info.IsDir() {
		return 0	
	}

	return info.Size()
}

/*
* CaculateFirmwareSize
* cacultae the firmware size.
*/
func CaculateFirmwareSize(pkgFileName, parFileName string) int64 {
	length := int64(0)
	pkgFileRoot := pkgFileName[0:strings.LastIndex(pkgFileName, "/")]
	
	packages := config.PackageFileLoad(pkgFileName)
	if packages == nil {
		return 0
	}
	
	partitions := config.GetDiskPartitionInfos(parFileName)
	if partitions == nil {
		return 0
	}
	
	length += int64(1932)
	for _, part := range partitions {
		name := part.Name
		fileName, ok := packages[name]
		if !ok {
			continue
		}
		
		pathFile := pkgFileRoot+ "/" + fileName
		length += CaculateFileSize(pathFile)
	}
	
	return length
}

func GenerateImage(fileName, parFileName string, packages map[string]string) error {

	return nil
}

/*
* GenerateRKImage 
* Generate RockChip linux image
*/
func GenerateRKImage(fileName, pkgFileName, parFileName string) error {
	length := int64(0)
	pkgFileRoot := pkgFileName[0:strings.LastIndex(pkgFileName, "/")]
	
	
	//1. create target image.
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}

	//2. load config file.
	packages := config.PackageFileLoad(pkgFileName)
	if packages == nil {
		f.Close()
		return errors.New("load package file failed")
	}
	partitions := config.GetDiskPartitionInfos(parFileName)
	if partitions == nil {
		f.Close()
		return errors.New("load partitions failed")
	}
	diskInfo := config.ParameterLoad(parFileName)
	if diskInfo == nil {
		f.Close()
		return errors.New("load diskInfo failed")
	}
	//3. initial the image header
	header := NewRKImageHeader(IMAGE_VERSION, IMAGE_VERSION, uint32(RK32_DEVICE))
	header.Size = uint16(102)
	offset := uint32(header.Size)

	//bootloader offset
	filePath, ok := packages["bootloader"]
	if !ok {
		length = 0
	}else{
		pathFile := pkgFileRoot + "/" + filePath
		length = CaculateFileSize(pathFile)
	}
	header.SetBootLoaderLocation(offset, uint32(length))

	flen := CaculateFirmwareSize(pkgFileName, parFileName)
	header.SetFirmwareLocation(offset +uint32(length), uint32(flen))
	
	//4. write RKImageHeader
	fmt.Printf("write RKImageHeader ... \r\n")
	err = WriteRKImageHeader(f, header)
	if err != nil {
		f.Close()
		return err
	}

	//5. write boot loader
	if length > 0 {
		pathFile := pkgFileRoot + "/" + filePath

		fb, err := os.Open(pathFile)
		if err != nil {
			f.Close()
			return errors.New("load "+ pathFile+ " failed")
		} 

		io.Copy(f, fb)
		fb.Close()
		
		offset += uint32(length)
	}

	
	//6. write firmware header.
	fmt.Printf("write firmware ...\r\n")
	imageHeaderOffset := offset
	h := NewImageHeader(diskInfo.MachineModel, diskInfo.Manufacturer)	
	err = WriteImageHeader(f, h)
	if err != nil {
		f.Close()
		return err
	}
	hsize := uint16(1932)
	offset = uint32(hsize)
	
	for _, part := range partitions {
		name := part.Name
		filePath, ok := packages[name]
		if !ok {
			continue
		}
		
		pathFile := pkgFileRoot+ "/" + filePath
		fw, err := os.Open(pathFile)
		if err != nil {
			f.Close()
			return err
		} 

		l, err := io.Copy(f, fw)
		if err != nil {
			f.Close()
			return err
		} 

		fmt.Printf("write %s: offset = %d  size =%d \r\n", name, offset, l)
		item := NewImageItem(name, filePath, uint32(offset), uint32(l), 1, int32(part.Offset))
		h.ImageItems[h.ItemCount] = *item
		h.ItemCount++
		fw.Close()
		
		offset += uint32(l)
	}
	
	_, err = f.Seek(int64(imageHeaderOffset), 0)
	if err != nil {
		f.Close()
		return err
	}
	
	err = WriteImageHeader(f, h)
	if err != nil {
		f.Close()
		return err
	}

	f.Close()
	
	//generate the md5 sum
	fmt.Printf("Generate the md5 sum ...\r\n")
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	md5 := md5.New()
	_, err = io.Copy(md5, file)
	if err != nil {
		file.Close()
		return err
	}
	md5Str := hex.EncodeToString(md5.Sum(nil))
	fmt.Printf("md5Str = %s\r\n",md5Str)

	_, err = file.Seek(0, 2)
	if err != nil {
		file.Close()
		return err
	}
	
	_, err = file.Write([]byte(md5Str))
	file.Close()
	
	return err
}
