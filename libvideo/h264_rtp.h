/*
 * h264_rtp.h
 *
 */

#ifndef H264_RTP_H
#define H264_RTP_H

#include <stdint.h>
#include "libv_log.h"
#include "libvideo.h"
/*
 * 此库暂不支持多线程
 * */

#ifdef __cplusplus
extern "C" {
#endif

typedef enum
{
    PACKET_MODE_SINGLE = 0, //单个包模式
    PACKET_MODE_NON_INTERLEAVED = 1,    //非交错方式

}PACKET_MODE_E;


typedef void * H264_RTP_DEC_TH;

H264_RTP_DEC_TH alloc_h264_rtp_dec_th();
void free_h264_rtp_dec_th(H264_RTP_DEC_TH  pdec_th);
//rtp_buf:包含rtp包头.返回值：获得包的个数.-1表示解析失败
int h264_dec_rtp(H264_RTP_DEC_TH  pdec_th, uint8_t *rtp_buf, int buf_len);

//函数返回数据的指针，没有数据则返回空。这个函数要一直调用，直到返回null为止，否则数据会被覆盖
//返回的指针或指针中的缓冲区不可以作为上层的存储缓冲区，如果需要保存一段时间，则需要重新申请新的缓冲区并拷贝。
//used_flag字段返回前会被置为0.
H264_NAL_DATA * h264_dec_rtp_get_nal_data(H264_RTP_DEC_TH  pdec_th);

typedef void * H264_RTP_ENC_TH;
/*
 *pload_type:负载类型；s_seq：起始序列号,使用随机数作为开始；
 *ssrc：同步源id;frame_rate：帧率，目前仅支持整数，后面可能会支持浮点数
 *mtu:单个ip包的最大限制，建议1500，此值包括ip包头，udp包头，rtp头的长度。
 * */
H264_RTP_ENC_TH * alloc_h264_rtp_enc_th(uint8_t pload_type, uint16_t s_seq, uint32_t ssrc, uint32_t mtu, PACKET_MODE_E pm);
void free_h264_rtp_enc_th(H264_RTP_ENC_TH  ph264dec_th);
//nal_buf：一个NAL单元数据，必须包含四字节[0,0,0,1]或三字节[0,0,1]的起始码
//timestamp:该NAL的时戳，需要上层传递，因为未必是连续的。
//成功返回：编码后rtp包个数；失败返回：-1；
int32_t h264_enc_rtp_one_nalu(H264_RTP_ENC_TH ph264dec_th, uint8_t *nal_buf, int32_t buf_len, uint32_t timestamp, uint32_t mark_flg);

//an_buf：一个access unit单元数据，必须包含四字节[0,0,0,1]或三字节[0,0,1]的起始码.
//timestamp:该NAL的时戳，需要上层传递，因为未必是连续的。
//成功返回：编码后rtp包个数；失败返回：-1；
int32_t h264_enc_rtp_access_unit(H264_RTP_ENC_TH ph264enc_th, uint8_t *an_buf, int32_t buf_len, uint32_t timestamp);

//函数返回数据的指针，没有数据则返回空。这个函数要一直调用，直到返回null为止，否则数据会被覆盖
//返回的指针或指针中的缓冲区不可以作为上层的存储缓冲区，如果需要保存一段时间，则需要重新申请新的缓冲区并拷贝。
//used_flag字段返回前会被置为0.
H264_RTP_PACKET * h264_enc_rtp_get_rtp_data(H264_RTP_ENC_TH  penc_th);

#ifdef __cplusplus
}
#endif



#endif

