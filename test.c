#include <stdio.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>

#define BLOCK_WRITE_LEN (16 * 1024)

typedef unsigned char BYTE;
typedef BYTE *PBYTE;
typedef unsigned short USHORT;
typedef unsigned int    UINT;
typedef unsigned int    DWORD;
typedef unsigned char UCHAR;
typedef unsigned short WCHAR;
typedef signed char CHAR;

typedef enum
{
    RKNONE_DEVICE=0,
    RK27_DEVICE=0x10,
    RKCAYMAN_DEVICE,
    RK28_DEVICE=0x20,
    RK281X_DEVICE,
    RKPANDA_DEVICE,
    RKNANO_DEVICE=0x30,
    RKSMART_DEVICE,
    RKCROWN_DEVICE=0x40,
    RK29_DEVICE=0x50,
    RK292X_DEVICE,
    RK30_DEVICE=0x60,
    RK30B_DEVICE,
    RK31_DEVICE=0x70,
    RK32_DEVICE=0x80
}ENUM_RKDEVICE_TYPE;

typedef struct
{
    USHORT  usYear;
    BYTE    ucMonth;
    BYTE    ucDay;
    BYTE    ucHour;
    BYTE    ucMinute;
    BYTE    ucSecond;
}STRUCT_RKTIME, *PSTRUCT_RKTIME;

typedef struct
{
    UINT uiTag;     //æ ‡å¿—ï¼Œå›ºå®šä¸º0x57 0x46 0x4B 0x52
    USHORT usSize;  //ç»“æž„ä½“å¤§å°
    DWORD  dwVersion;   //Image æ–‡ä»¶ç‰ˆæœ¬
    DWORD  dwMergeVersion;  //æ‰“åŒ…å·¥å…·ç‰ˆæœ¬
    STRUCT_RKTIME stReleaseTime;    //ç”Ÿæˆæ—¶é—´
    ENUM_RKDEVICE_TYPE emSupportChip;   //ä½¿ç”¨èŠ¯ç‰‡
    DWORD  dwBootOffset;    //Bootåç§»
    DWORD  dwBootSize;  //Bootå¤§å°
    DWORD  dwFWOffset;  //å›ºä»¶åç§»
    DWORD  dwFWSize;    //å›ºä»¶å¤§å°
    BYTE   reserved[61];    //é¢„ç•™ç©ºé—´ï¼Œç”¨äºŽå­˜æ”¾ä¸åŒå›ºä»¶ç‰¹å¾
}STRUCT_RKIMAGE_HEAD,*PSTRUCT_RKIMAGE_HEAD;

typedef struct tagRKIMAGE_ITEM
{
    char name[32];
    char file[64];
    unsigned int offset;
    unsigned int flash_offset;
    unsigned int usespace;
    unsigned int size;
}RKIMAGE_ITEM, *PRKIMAGE_ITEM;

typedef struct tagRKIMAGE_HDR
{
    unsigned int tag;
    unsigned int size;
    char machine_model[64];
    char manufacturer[60];
    unsigned int version;
    int item_count;
    RKIMAGE_ITEM item[16];
}RKIMAGE_HDR, *PRKIMAGE_HDR;




static void display_head(PSTRUCT_RKIMAGE_HEAD pHead){
    printf("uiTag = %x.\n", pHead->uiTag);
    printf("usSize = %x.\n", pHead->usSize);
    printf("dwVersion = %x.\n", pHead->dwVersion);
    unsigned int btMajor = ((pHead->dwVersion) & 0XFF000000) >> 24;
    unsigned int btMinor = ((pHead->dwVersion) & 0X00FF0000) >> 16;
    unsigned int usSmall = ((pHead->dwVersion) & 0x0000FFFF);
    printf("btMajor = %x, btMinor = %x, usSmall = %02x.\n", btMajor, btMinor, usSmall);
    printf("emSupportChip = %x.\n", pHead->emSupportChip);
    printf("dwBootOffset = %x.\n", pHead->dwBootOffset);
    printf("dwBootSize = %x.\n", pHead->dwBootSize);
    printf("dwFWOffset = %x.\n", pHead->dwFWOffset);
    printf("dwFWSize = %d.\n", pHead->dwFWSize);
}

static void display_item(PRKIMAGE_ITEM pitem) {
    printf("name = %s\r\n", pitem->name);
    printf("file = %s \r\n", pitem->file);
    printf("offset = %d \r\n", pitem->offset);
    printf("flash_offset = %d \r\n", pitem->flash_offset);
    printf("usespace = %d \r\n", pitem->usespace);
    printf("size = %d \r\n", pitem->size);
}

static void display_hdr(PRKIMAGE_HDR phdr) {

    printf("tag = %d \r\n", phdr->tag);
    printf("size = %d \r\n", phdr->size);
    printf("machine_model = %s \r\n", phdr->machine_model);
    printf("manufacturer = %s \r\n", phdr->manufacturer);
    printf("version = %d \r\n", phdr->version);
    printf("item = %d.\n", phdr->item_count);
    for(int i = 0; i < phdr->item_count; i++){
        printf("================================================");
        display_item(&(phdr->item[i]));
    }
}

int main(){
	STRUCT_RKIMAGE_HEAD rkimage_head;
	RKIMAGE_HDR hdr;
	RKIMAGE_ITEM* item = NULL;
	PRKIMAGE_HDR phdr = &hdr;	
	char data_buf[BLOCK_WRITE_LEN] = {0};
	long long src_remain, dest_remain;
	int src_step, dest_step;
	long long read_count, write_count;
	
	int fd = open("update.img", O_RDONLY);
    if (fd < 0) {
        printf("Can't open update.img\n");
        return -2;
    }
    
    if (read(fd, &rkimage_head, sizeof(STRUCT_RKIMAGE_HEAD)) != sizeof(STRUCT_RKIMAGE_HEAD)) {
        printf("Can't read %s\n(%s)\n", "update.img", strerror(errno));
        close(fd);
        return -2;
    }
	display_head(&rkimage_head);
	
	if (lseek64(fd, rkimage_head.dwFWOffset, SEEK_SET) == -1) {
        printf("lseek failed.\n");
        return -2;
    }
    
    if (read(fd, phdr, sizeof(RKIMAGE_HDR)) != sizeof(RKIMAGE_HDR)) {
        printf("Can't read (%s)\n", strerror(errno));
        close(fd);
        return -2;
    }
    
    if (phdr->tag != 0x46414B52) {
        printf("tag: %x\n", phdr->tag);
        printf("Invalid image\n");
        return -3;
    }
    
    display_hdr(phdr);

	int i  = 0;
	for( i = 0 ; i < phdr->item_count; i++){
		int fd_dest;
		unsigned int offset;
	    unsigned int size;
	    
		item = &phdr->item[i];
		
		offset = item->offset;
		size = item->size;
		dest_remain = src_remain = size;
		dest_step = src_step = BLOCK_WRITE_LEN;
		
		fd_dest = open(item->name, O_RDWR | O_CREAT, 0644);
    	if (fd_dest < 0) {
        	printf("Can't open %s\n", item->name);
        	return -2;
    	}
    	
		if (lseek64(fd, offset, SEEK_SET) == -1) {
        	printf("lseek failed.\n");
        	return -2;
    	}
    	
    	while (src_remain > 0 && dest_remain > 0) {
        	memset(data_buf, 0, BLOCK_WRITE_LEN);
        	read_count = src_remain>src_step?src_step:src_remain;

        	if (read(fd, data_buf, read_count) != read_count) {
         	   close(fd_dest);
        	   close(fd);
         	   printf("Read failed(%s):(%s:%d)\n", strerror(errno), __func__, __LINE__);
         	   return -2;
        	}
        	
        	src_remain -= read_count;
        	write_count = dest_remain>dest_step?dest_step:dest_remain;
        	
        	if (write(fd_dest, data_buf, write_count) != write_count) {
            	close(fd_dest);
            	close(fd);
            	printf("Write failed(%s):(%s:%d)\n", strerror(errno), __func__, __LINE__);
            	return -2;
        	}
        	dest_remain -= write_count;
        }
        
        fsync(fd_dest);
        close(fd_dest);
	}
    close(fd);
    
    return 0;
}
