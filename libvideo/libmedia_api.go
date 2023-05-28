package libvideo

/*
#cgo CFLAGS: -I./
#cgo LDFLAGS: -L/usr/lib/testVideo -L/usr/lib -L/usr/local/lib -L/usr/lib64 -lmedia -lavformat -lavcodec -lavutil -lx264 -lopenh264 -lswscale -liconv -lm
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "libvideo.h"
#include "h264_decoder.h"
#include "h264_encoder.h"
#include "h264_rtp.h"
#include "libavutil/pixfmt.h"
#include "video_mix.h"

void C_VideoCodec_Init(int logLevel, char* logFile, int isPrint)
{
	if (logLevel < E_LIBV_LOG_ERROR)
	{
		logLevel = E_LIBV_LOG_ERROR;
	}
	else if (logLevel > E_LIBV_LOG_MAX)
	{
		logLevel = E_LIBV_LOG_MAX;
	}

	libv_init(NULL, logLevel, logFile, isPrint);
}

H264_DECODER_TH *C_Decode_Dec_Th_Alloc(int iDecodeId)
{
    if(iDecodeId > H264_DECODER_ID_MAX || iDecodeId < 0)
    {
        return NULL;
    }


    return alloc_decoder_by_id((H264_DECODER_ID)iDecodeId);
}


void C_Decode_Dec_Th_Free(H264_DECODER_TH *VideoDecIndex)
{
    if(!VideoDecIndex)
    {

        return;
    }

    free_decoder(VideoDecIndex);
}

//mix 相关函数
typedef struct VIDEO_MIX_WARP{
    VIDEO_MIX_TH*mix;
    FRAME_PIC_DATA* pPic[32];
    int MixNum;
    int width;
    int height;
	FRAME_PIC_DATA *pOut;
}VIDEO_MIX_WARP;

int GetMixNum(int mixNum){
    if(mixNum ==5){
        return 6;
    }
    if(mixNum ==7){
        return 8;
    }
    if(mixNum>=10 &&mixNum<=12){
        return 13;
    }
    if(mixNum>=14 &&mixNum<=15){
        return 16;
    }
    if(mixNum>=17 &&mixNum<=24){
        return 25;
    }
    return mixNum;
}

void C_Mix_Warp_Free(VIDEO_MIX_WARP*warp);

VIDEO_MIX_WARP*C_Mix_Warp_Alloc(int mixNum,int width,int height,int fmt){
    if(mixNum<=0 && mixNum>=26){
        return NULL;
    }
	mixNum = GetMixNum(mixNum);
    if(width<0 || height<0){
        return NULL;
    }
    VIDEO_MIX_WARP*warp = malloc(sizeof(VIDEO_MIX_WARP));
    if(warp == NULL){
        return NULL;
    }
    memset(warp,0,sizeof(VIDEO_MIX_WARP));
    VIDEO_MIX_TH*mix = alloc_video_mix_th(mixNum,width,height,fmt);
    if(mix == NULL){
        free(warp);
        return NULL;
    }
    warp->mix = mix;
    warp->height = height;
    warp->width = width;
    warp->MixNum = mixNum;
	warp->pOut = libv_alloc_frame_pic_mem(0,warp->width,warp->height);
	if(warp->pOut == NULL){
		C_Mix_Warp_Free(warp);
        return NULL;
	}
    return warp;
}

void C_Mix_Warp_Free(VIDEO_MIX_WARP*warp){
    if(warp == NULL){
        return;
    }
    free_video_mix_th(warp->mix);
	if(warp->pOut!=NULL){
		libv_free_frame_pic(warp->pOut);
	}
    free(warp);
}

//这个p应该是解码出来noline的那种，index表示位置
int C_Mix_Warp_AddPeerPic(VIDEO_MIX_WARP*warp,FRAME_PIC_DATA*p,int index){
    if(warp == NULL|| index <0){
        return -1;
    }
    //调整位置
    (warp->pPic)[index] = p;
	return 0;
}

int C_Reset_Mix_Num(VIDEO_MIX_WARP*mix,int mixNum){
    mixNum = GetMixNum(mixNum);
    int ret = Reset_video_mix(mix->mix,mixNum);
    if(ret == 0){
        mix->MixNum = mixNum;
    }
    return ret;
}

//成功返回一个编码后的数据,不需要自己管理out的你村

FRAME_PIC_DATA* C_Video_MixWarp_Proc(VIDEO_MIX_WARP*warp){
    if(warp == NULL){
        return NULL;
    }
    FRAME_PIC_DATA *out=warp->pOut;
    if(out == NULL){
        return NULL;
    }
    int i = 0;
    int ret = video_mix_pic(warp->mix,warp->pPic,warp->MixNum,out);
    if(ret == 0){
        return out;
    }
	return NULL;
}


//解码相关函数

int C_VideoDecoderProcSimple(H264_DECODER_TH* decoder, int *width, int*height, char *dstData, char* newData, int newDataLen)
{
	int iGotFrame = 0;
	H264_NAL_DATA  NalDataTmp;
	H264_NAL_DATA* pNalDataTmp = &NalDataTmp;
	FRAME_PIC_DATA Pic;
	FRAME_PIC_DATA *pPic = &Pic;
	if (newDataLen > 0 && newData != NULL)
	{
		//memcpy(pNalDataTmp->data_buf, newData, newDataLen);
		pNalDataTmp->data_buf = newData;
		pNalDataTmp->data_size = newDataLen;
	}

	int iDecLen = decoder->decode_video(decoder, pPic, &iGotFrame, pNalDataTmp);
	int iWidHeig = 0;
	if(iGotFrame && iDecLen >= 0)
	{
		iWidHeig = pPic->width*pPic->height;
		if (iWidHeig*3/2 <= 4096*1024)
		{
			*width = pPic->width;
			*height = pPic->height;

			int j=0;
			for(j=0; j<pPic->height; j++)
				memcpy(dstData + j * pPic->width, pPic->data[0] + j * pPic->linesize[0], pPic->width);
			for(j=0; j<pPic->height/2; j++)
				memcpy(dstData + (iWidHeig + j * pPic->width/2), pPic->data[1] + j * pPic->linesize[1], pPic->width/2);
			for(j=0; j<pPic->height/2; j++)
				memcpy(dstData+ (5*iWidHeig/4 + j* pPic->width/2), pPic->data[2] + j * pPic->linesize[2], pPic->width/2);
		}
	}else{
		return -1;
	}

	return iWidHeig*3/2;
}

//just for test
int C_Video_GetFrameData(FRAME_PIC_DATA* pPic,char** dstData){
	if(*dstData == NULL){
		return -1;
	}
	*dstData = pPic->data[0];
	return pPic->width*pPic->height;;
}

//即pPic中的数据不进行排序,其实主要作用于mix,这里的pic其实没有复制内存的
int C_VideoDecoderProcFrameNoLine(H264_DECODER_TH* decoder, FRAME_PIC_DATA* pPic, char* newData, int newDataLen)
{
	H264_NAL_DATA  NalDataTmp;
    H264_NAL_DATA* pNalDataTmp = &NalDataTmp;
    int iGotFrame = 0;
	if (newData == NULL){
		return -1;
	}
    if (newDataLen > 0 && newData != NULL)
    {
        pNalDataTmp->data_buf = newData;
        pNalDataTmp->data_size = newDataLen;
    }

    int iDecLen = decoder->decode_video(decoder, pPic, &iGotFrame, pNalDataTmp);
    int iWidHeig = 0;
    if(iGotFrame && iDecLen >= 0)
    {
        iWidHeig = pPic->width*pPic->height;
		if (iWidHeig*3/2 > 4096*1024){
			return -2;
		}
    }else{
		return -1;
	}
    return iWidHeig*3/2;
}


//将数据排序
int C_VideoDecoderProcFrameLine(FRAME_PIC_DATA* pPic, char* dstData){
	int iWidHeig = pPic->width*pPic->height;
	if(iWidHeig<=0){
		return -1;
	}
	if(dstData == NULL){
		return -1;
	}
	if (iWidHeig*3/2 <= 4096*1024)
	{
		int j=0;
		for(j=0; j<pPic->height; j++)
			memcpy(dstData + j * pPic->width, pPic->data[0] + j * pPic->linesize[0], pPic->width);
		for(j=0; j<pPic->height/2; j++)
			memcpy(dstData + (iWidHeig + j * pPic->width/2), pPic->data[1] + j * pPic->linesize[1], pPic->width/2);
		for(j=0; j<pPic->height/2; j++)
			memcpy(dstData+ (5*iWidHeig/4 + j* pPic->width/2), pPic->data[2] + j * pPic->linesize[2], pPic->width/2);

		return  iWidHeig*3/2;
	}
	return  -2;
}


int C_VideoDecoderProcFrame(H264_DECODER_TH* decoder, FRAME_PIC_DATA* pPic, char* newData, int newDataLen)
{
	H264_NAL_DATA  NalDataTmp;
    H264_NAL_DATA* pNalDataTmp = &NalDataTmp;
    int iGotFrame = 0;
	if (newData == NULL){
		return -1;
	}
    if (newDataLen > 0 && newData != NULL)
    {
        pNalDataTmp->data_buf = newData;
        pNalDataTmp->data_size = newDataLen;
    }

    int iDecLen = decoder->decode_video(decoder, pPic, &iGotFrame, pNalDataTmp);
    int iWidHeig = 0;
    if(iGotFrame && iDecLen >= 0)
    {
        iWidHeig = pPic->width*pPic->height;
        if (iWidHeig*3/2 <= 4096*1024)
        {
            unsigned char *dstData = malloc(iWidHeig*3/2);
            if(dstData == NULL){
                return -1;
            }
            int j=0;
            for(j=0; j<pPic->height; j++)
                memcpy(dstData + j * pPic->width, pPic->data[0] + j * pPic->linesize[0], pPic->width);
            for(j=0; j<pPic->height/2; j++)
                memcpy(dstData + (iWidHeig + j * pPic->width/2), pPic->data[1] + j * pPic->linesize[1], pPic->width/2);
            for(j=0; j<pPic->height/2; j++)
                memcpy(dstData+ (5*iWidHeig/4 + j* pPic->width/2), pPic->data[2] + j * pPic->linesize[2], pPic->width/2);

            pPic->data[0] = dstData;
            pPic->_alloc_flg = 1;
        }
    }else{
		return -1;
	}
    return iWidHeig*3/2;
}

FRAME_PIC_DATA* C_VideoDecodeFrameCreate(int fmt,int w,int h){
	FRAME_PIC_DATA*data = libv_alloc_frame_pic(fmt,w,h);	//不申请内存
	if(data == NULL){
		return NULL;
	}
	return data;
}


FRAME_PIC_DATA* C_VideoDecodeFrameCopy(FRAME_PIC_DATA* src){
	FRAME_PIC_DATA *data = libv_copy_new_frame_pic(src);	//申请内存的
	if(data == NULL){
		return NULL;
	}
	return data;
}

int C_VideoDecodeFrameDelete(FRAME_PIC_DATA* pPic){
	if(pPic == NULL){
		return -1;
	}
	libv_free_frame_pic(pPic);
	return 0;
}

int C_VideoDecoderProc(H264_DECODER_TH* decoder, H264_NAL_DATA* pNalDataTmp, int *width, int*height, char *dstData, char* newData, int newDataLen)
{
	int iGotFrame = 0;
	FRAME_PIC_DATA *pPic = libv_alloc_frame_pic(PIC_FMT_YUV_I420, 0, 0); //申请一个picture单元
	if (pNalDataTmp->nal_type == NALU_TYPE_IDR && newDataLen > 0 && newData != NULL)
	{
		memcpy(pNalDataTmp->data_buf, newData, newDataLen);
		pNalDataTmp->data_size = newDataLen;
	}

	int iDecLen = decoder->decode_video(decoder, pPic, &iGotFrame, pNalDataTmp);
	int iWidHeig = 0;
	if(iGotFrame && iDecLen >= 0)
	{
		iWidHeig = pPic->width*pPic->height;
		if (iWidHeig*3/2 <= 4096*1024)
		{
			*width = pPic->width;
			*height = pPic->height;

			int j=0;
			for(j=0; j<pPic->height; j++)
				memcpy(dstData + j * pPic->width, pPic->data[0] + j * pPic->linesize[0], pPic->width);
			for(j=0; j<pPic->height/2; j++)
				memcpy(dstData + (iWidHeig + j * pPic->width/2), pPic->data[1] + j * pPic->linesize[1], pPic->width/2);
			for(j=0; j<pPic->height/2; j++)
				memcpy(dstData+ (5*iWidHeig/4 + j* pPic->width/2), pPic->data[2] + j * pPic->linesize[2], pPic->width/2);
		}
	}

	libv_free_frame_pic(pPic);

	return iWidHeig*3/2;
}


//编码相关函数
H264_RTP_ENC_TH C_alloc_video_rtp_enc_th(char iPayLoad, short iSseq, int iSsrc, int iMtu, int packetMode)
{
    if(iPayLoad < 0 || iSseq < 0 || iSsrc < 0 || iMtu < 0)
    {
        return NULL;
    }

	return alloc_h264_rtp_enc_th(iPayLoad, iSseq, iSsrc, iMtu, packetMode);
}

H264_ENCODER_TH* C_alloc_encoder(int iEncodeId, int width, int height, int rate, int bitrate, int sliceMode)
{
	ENCODE_PARAMETER pParam;
	memset(&pParam, 0, sizeof(pParam));
	pParam.complexity_mode = 1;
	pParam.frame_rate = rate;
	pParam.height = height;
	pParam.width = width;

	if (height*width < 1280*720 / 2)
	{
	    pParam.multiple_thread_id = 2;
	}
	else
	{
	    pParam.multiple_thread_id = 3;
	}

	pParam.pic_format = PIC_FMT_YUV_I420;
	pParam.slice_cfg.max_slice_size = 1400 - 100;
	pParam.slice_mode = (H264_SLICE_MODE_E)sliceMode;  //H264_SM_AUTO_SLICE ; //H264_SM_DYN_SLICE;//
	pParam.bitrate = bitrate*1024;//
	pParam.max_bitrate = bitrate*1024;

	return alloc_encoder_by_id((H264_ENCODER_ID)iEncodeId, &pParam);
}

int C_encode_frame(H264_ENCODER_TH* encoder, int width, int height, int timestamp, char* yuvBuf, LAYER_BS_INFO* layerBsInfo)
{
	FRAME_PIC_DATA pic;
	memset(&pic, 0, sizeof(pic));

	pic.pic_format = PIC_FMT_YUV_I420;
	pic.timestamp = timestamp;
	pic.height = height;
	pic.width = width;

	pic.linesize[0] = width;//y  宽*高
	pic.linesize[1] = width/2;//u   y/2
	pic.linesize[2] = width/2;//v   y/2

	pic.data[0] = (uint8_t *)yuvBuf;
	pic.data[1] = (uint8_t *)yuvBuf + width*height;
	pic.data[2] = pic.data[1] + width*height/4;

	memset(layerBsInfo, 0, sizeof(LAYER_BS_INFO));

	return encoder->encode_frame(encoder, &pic, layerBsInfo);
}

void C_force_intra_frame(H264_ENCODER_TH* encoder)
{
	encoder->force_intra_frame(encoder, 1);
}

*/
import "C"

import (
	"McuConference/libvideo_codec/av"
	"time"
	"unsafe"
)

var RESOLUTION_4K = 4096 * 2160 * 3 / 2

func LVC_Init(logLevel int, logfile string, isPrint bool) {
	var C_isPrint C.int
	if isPrint {
		C_isPrint = 1
	} else {
		C_isPrint = 0
	}

	C_logFile := C.CString(logfile)
	defer C.free(unsafe.Pointer(C_logFile))

	C.C_VideoCodec_Init(C.int(logLevel), C_logFile, C_isPrint)
}

/**********************************************************************************************************************/
/***************************************   以下方法为将nal解码为yuv  ****************************************************/
/**********************************************************************************************************************/
const (
	DEC_ENC_ID_H264_FFM      = iota // ffmpeg
	DEC_ENC_ID_H264_OPENH264        // openh264
	DEC_ENC_ID_H265_HEVC            // H265解码
	DEC_ENC_ID_VP8                  // VP8解码
)

type LVC_VideoDecoderST struct {
	Decoder   *C.H264_DECODER_TH
	DecoderId int
	Width     int
	Height    int
	Fmt       int
}

func LVC_CreatDecoder(DECODER_ID int) *LVC_VideoDecoderST {
	videoDecoder := new(LVC_VideoDecoderST)
	videoDecoder.DecoderId = DECODER_ID
	{
		videoDecoder.Decoder = C.C_Decode_Dec_Th_Alloc(C.int(DECODER_ID))
		if videoDecoder.Decoder == nil {
			return nil
		}
	}

	return videoDecoder
}

func LVC_FreeDecoder(videoDecoder *LVC_VideoDecoderST) {
	{
		C.C_Decode_Dec_Th_Free(videoDecoder.Decoder)
	}
}

// h264解码的时候，newData的数据是将， sps  pps  sei idr合成一个数据包的数据，一个个传进去解码会报no freame错误
// Vp8的话，vpx通过newData传入数据
//func LVC_VideoDecoderProc(videoDecoder *LVC_VideoDecoderST, pNalDataSt *C.H264_NAL_DATA, newData []byte) (YuvData [][]byte, ok bool) {
//
//	{
//		if videoDecoder == nil || pNalDataSt == nil {
//			return nil, false
//		}
//
//		yuvData := make([]byte, RESOLUTION_4K)
//		width := 0
//		height := 0
//		dataLen := 0
//		if newData == nil {
//			dataLen = int(C.C_VideoDecoderProc(videoDecoder.Decoder, pNalDataSt, (*C.int)(unsafe.Pointer(&width)),
//				(*C.int)(unsafe.Pointer(&height)), (*C.char)(unsafe.Pointer(&yuvData[0])), nil, 0))
//		} else {
//			dataLen = int(C.C_VideoDecoderProc(videoDecoder.Decoder, pNalDataSt, (*C.int)(unsafe.Pointer(&width)),
//				(*C.int)(unsafe.Pointer(&height)), (*C.char)(unsafe.Pointer(&yuvData[0])),
//				(*C.char)(unsafe.Pointer(&newData[0])), C.int(len(newData))))
//		}
//
//		if dataLen > 0 {
//			ok = true
//			YuvData = append(YuvData, yuvData[:dataLen])
//			videoDecoder.Width = width
//			videoDecoder.Height = height
//		}
//	}
//
//	return
//}

type LvcFramePicDataSt struct {
	Data *C.FRAME_PIC_DATA
}

func LvcFramePicData_Create(f, w, h int) *LvcFramePicDataSt {
	l := LvcFramePicDataSt{Data: nil}
	l.Data = C.C_VideoDecodeFrameCreate(C.int(f), C.int(w), C.int(h))
	return &l
}

func LvcFramePicData_Delete(l *LvcFramePicDataSt) {
	C.C_VideoDecodeFrameDelete(l.Data)
}

func Lvc_Video_GetFrameData(l *LvcFramePicDataSt) int {
	d := int(C.C_Video_GetFrameData(l.Data, nil))
	return d
}

func LVC_VideoDecoderProcFrame(videoDecoder *LVC_VideoDecoderST, pData *LvcFramePicDataSt, newData []byte) (ok bool) {

	{
		if videoDecoder == nil || pData == nil {
			return false
		}
		dataLen := 0
		if newData == nil {
			dataLen = int(C.C_VideoDecoderProcFrame(videoDecoder.Decoder, pData.Data, nil, 0))
		} else {
			dataLen = int(C.C_VideoDecoderProcFrame(videoDecoder.Decoder, pData.Data, (*C.char)(unsafe.Pointer(&newData[0])), C.int(len(newData))))
		}
		if dataLen > 0 {
			ok = true
		}
	}

	return
}

func LVC_VideoDecoderProcFrameNoLine(videoDecoder *LVC_VideoDecoderST, pData *LvcFramePicDataSt, newData []byte) (ok bool) {

	{
		if videoDecoder == nil || pData == nil {
			return false
		}
		dataLen := 0
		if newData == nil {
			dataLen = int(C.C_VideoDecoderProcFrameNoLine(videoDecoder.Decoder, pData.Data, nil, 0))
		} else {
			dataLen = int(C.C_VideoDecoderProcFrameNoLine(videoDecoder.Decoder, pData.Data, (*C.char)(unsafe.Pointer(&newData[0])), C.int(len(newData))))
		}
		if dataLen > 0 {
			ok = true
		}
	}

	return
}

func LVC_VideoDecoderProcFrameLine(pData *LvcFramePicDataSt) ([]byte, bool) {

	{
		if pData == nil {
			return nil, false
		}
		yuvData := make([]byte, RESOLUTION_4K)
		dataLen := 0
		dataLen = int(C.C_VideoDecoderProcFrameLine(pData.Data, (*C.char)(unsafe.Pointer(&yuvData[0]))))
		if dataLen > 0 {
			YuvData := yuvData[:dataLen]
			return YuvData, true
		} else {
			return nil, false
		}
	}

}

func LVC_VideoDecoderProcSimple2(videoDecoder *LVC_VideoDecoderST, newData []byte) (YuvData []byte, ok bool) {

	{
		if videoDecoder == nil {
			return nil, false
		}

		yuvData := make([]byte, RESOLUTION_4K)
		width := 0
		height := 0
		dataLen := 0
		if newData == nil {
			dataLen = int(C.C_VideoDecoderProcSimple(videoDecoder.Decoder, (*C.int)(unsafe.Pointer(&width)),
				(*C.int)(unsafe.Pointer(&height)), (*C.char)(unsafe.Pointer(&yuvData[0])), nil, 0))
		} else {
			dataLen = int(C.C_VideoDecoderProcSimple(videoDecoder.Decoder, (*C.int)(unsafe.Pointer(&width)),
				(*C.int)(unsafe.Pointer(&height)), (*C.char)(unsafe.Pointer(&yuvData[0])),
				(*C.char)(unsafe.Pointer(&newData[0])), C.int(len(newData))))
		}

		if dataLen > 0 {
			ok = true
			YuvData = yuvData[:dataLen]
			videoDecoder.Width = width
			videoDecoder.Height = height
		} else {

		}
	}

	return
}

/**********************************************************************************************************************/
/**************************************************混码*****************************************************************/
/**********************************************************************************************************************/
type LVC_VideoMixST struct {
	MixHandle *C.VIDEO_MIX_WARP
	Width     int
	Height    int
	Fmt       int
}

func LVC_CreateMixHandle(mixNum, width, height int) *LVC_VideoMixST {
	mixHandle := C.C_Mix_Warp_Alloc(C.int(mixNum), C.int(width), C.int(height), C.int(0))
	if mixHandle == nil {
		return nil
	}
	mix := new(LVC_VideoMixST)
	mix.Fmt = 0
	mix.Height = height
	mix.Width = width
	mix.MixHandle = mixHandle
	return mix
}

func LVC_ResetMixHandle(mix *LVC_VideoMixST, num int) int {
	if mix == nil {
		return -1
	}
	if mix.MixHandle == nil {
		return -1
	}
	C.C_Reset_Mix_Num(mix.MixHandle, C.int(num))
	return 0
}

func LVC_DeleteMixHandle(mix *LVC_VideoMixST) int {
	if mix == nil {
		return -1
	}
	if mix.MixHandle == nil {
		return -1
	}
	C.C_Mix_Warp_Free(mix.MixHandle)
	return 0
}

//真正的混码函数
func LVC_ProcMixHandle(mix *LVC_VideoMixST) *LvcFramePicDataSt {
	if mix == nil {
		return nil
	}
	if mix.MixHandle == nil {
		return nil
	}
	l := LvcFramePicDataSt{Data: nil}
	l.Data = C.C_Video_MixWarp_Proc(mix.MixHandle)
	if l.Data == nil {
		return nil
	}
	return &l
}

func LVC_AddMixPicture(mix *LVC_VideoMixST, pic *LvcFramePicDataSt, index int) int {
	if mix == nil {
		return -1
	}
	if mix.MixHandle == nil {
		return -1
	}
	C.C_Mix_Warp_AddPeerPic(mix.MixHandle, pic.Data, C.int(index))
	return 0
}

/**********************************************************************************************************************/
/***************************************   以下方法为将yuv编码nal  ****************************************************/
/**********************************************************************************************************************/

//typedef struct __LAYER_BS_INFO
//{
//uint8_t tmp_id;
//FRAME_TYPE frame_type;       //帧的类型，和nal的类型是不同的，指明是I,P,B帧等
//
//int32_t  nal_count;                  // Count number of nal_data already
//H264_NAL_DATA  nal_data[MAX_NAL_UNITS_COUNT_IN_LAYER];
//} LAYER_BS_INFO;

type LVC_VideoEncoderST struct {
	Encoder     *C.H264_ENCODER_TH
	LayerBsInfo C.LAYER_BS_INFO
	EncoderId   int
	Width       int
	Height      int
	Rate        int
	Bitrate     int
	SliceMode   int

	SendIdrTimes    int
	SendIDRSecond   int
	SendNonIDRCount int

	Firsttimestamp int
	Timestamp      int
	Lasttime       time.Duration
}

func LVC_CreatEncoder(ENCODER_ID, width, height, rate, bitrate, sliceMode int) *LVC_VideoEncoderST {
	videoEncoder := new(LVC_VideoEncoderST)
	videoEncoder.EncoderId = ENCODER_ID
	videoEncoder.Width = width
	videoEncoder.Height = height
	videoEncoder.Rate = rate
	videoEncoder.Bitrate = bitrate
	videoEncoder.SliceMode = sliceMode
	videoEncoder.SendIDRSecond = 4

	{
		videoEncoder.Encoder = C.C_alloc_encoder(C.int(ENCODER_ID), C.int(width), C.int(height), C.int(rate),
			C.int(bitrate), C.int(sliceMode))
		if videoEncoder.Encoder == nil {
			return nil
		}
	}
	return videoEncoder
}

func LVC_FreeEncoder(videoEncoder *LVC_VideoEncoderST) {
	if videoEncoder == nil {
		return
	}

	{
		C.free_encoder(videoEncoder.Encoder)
	}
}

func LVC_VideoEncoderProc(videoEncoder *LVC_VideoEncoderST, timestamp int, yuvBuf []byte) (av.Packet, bool) {
	keyFramFlag := 0
	if videoEncoder.SendIdrTimes == 0 {
		keyFramFlag = 1
		videoEncoder.SendIdrTimes++
	}

	videoEncoder.SendNonIDRCount++
	if videoEncoder.SendIdrTimes < 6 {
		if videoEncoder.SendNonIDRCount >= videoEncoder.Rate {
			videoEncoder.SendIdrTimes++
			videoEncoder.SendNonIDRCount = 0
			keyFramFlag = 1
		}
	} else {
		if videoEncoder.SendNonIDRCount >= videoEncoder.Rate*videoEncoder.SendIDRSecond {
			keyFramFlag = 1
			videoEncoder.SendNonIDRCount = 0
		}
	}

	{
		var pkt av.Packet
		ok := false
		if keyFramFlag == 1 {
			C.C_force_intra_frame(videoEncoder.Encoder)
		}

		ret := int(C.C_encode_frame(videoEncoder.Encoder, C.int(videoEncoder.Width), C.int(videoEncoder.Height), C.int(timestamp),
			(*C.char)(unsafe.Pointer(&yuvBuf[0])), (*C.LAYER_BS_INFO)(unsafe.Pointer(&(videoEncoder.LayerBsInfo)))))

		if ret == 0 {
			pkt.Idx = 0
			for i := 0; i < int(videoEncoder.LayerBsInfo.nal_count); i++ {
				slice := C.GoBytes(unsafe.Pointer(videoEncoder.LayerBsInfo.nal_data[i].data_buf),
					C.int(videoEncoder.LayerBsInfo.nal_data[i].data_size))
				pkt.Data = append(pkt.Data, slice...)
				pkt.Datas = append(pkt.Datas, slice)
				if int(videoEncoder.LayerBsInfo.nal_data[i].nal_type) == int(NALU_TYPE_IDR) {
					pkt.IsKeyFrame = true
				}
			}

			/*		file, _ := os.OpenFile("poc111.h264", os.O_CREATE|os.O_APPEND|os.O_WRONLY ,0666)
					file.Write(pkt.Data)
					file.Close()*/

			videoEncoder.Timestamp = int(videoEncoder.LayerBsInfo.nal_data[0].timestamp)
			if videoEncoder.Firsttimestamp == 0 {
				videoEncoder.Firsttimestamp = videoEncoder.Timestamp
			}
			videoEncoder.Timestamp -= videoEncoder.Firsttimestamp
			pkt.Time = time.Duration(videoEncoder.Timestamp) * time.Second / time.Duration(90000)
			if pkt.Time < videoEncoder.Lasttime || pkt.Time-videoEncoder.Lasttime > time.Minute*30 {
				ok = false
				return pkt, ok
			}
			videoEncoder.Lasttime = pkt.Time
			ok = true
		}
		return pkt, ok
	}
}
