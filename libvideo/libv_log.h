/*
 * libv_log.h
 *
 */

#ifndef LIBV_LOG_H_
#define LIBV_LOG_H_
#include <stdarg.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum __LIBV_LOG_LEVEL_E
{
    E_LIBV_LOG_ERROR,
    E_LIBV_LOG_WARING,
    E_LIBV_LOG_INFO,
    E_LIBV_LOG_DEBUG,
    E_LIBV_LOG_MAX,
}LIBV_LOG_LEVEL_E;

//ȫ����־����
#define input_error                   libv_log(E_LIBV_LOG_ERROR, __FILE__, __LINE__, "The funtion input is error!");
#define elog(...)                   libv_log(E_LIBV_LOG_ERROR, __FILE__, __LINE__, __VA_ARGS__)
#define wlog(...)                   libv_log(E_LIBV_LOG_WARING, __FILE__, __LINE__, __VA_ARGS__)
#define ilog(...)                   libv_log(E_LIBV_LOG_INFO, __FILE__, __LINE__, __VA_ARGS__)
#define dlog(...)                   libv_log(E_LIBV_LOG_DEBUG, __FILE__, __LINE__, __VA_ARGS__)

typedef void ( * LIBV_LOG_FUN )(LIBV_LOG_LEVEL_E level, const char* file_name, int line_no, const char* format, va_list vl);

void libv_log(LIBV_LOG_LEVEL_E level, const char* file_name, int line_no, const char* format, ...);

/*
 ���log_cbfun��Ϊ�գ���ص��ϲ����־����������д�뵽�ļ�log_file��;printf_flg:�Ƿ��ӡ��ǰ̨
 ����Ƕ��̵߳��ã���Ҫ�ϲ�ص�������־���ײ㲻֧�ֶ��̴߳�����־��
 * */
void libv_log_set(LIBV_LOG_FUN log_cbfun, LIBV_LOG_LEVEL_E level, const char* log_file, int printf_flg);

void ffm_av_log_callback(void* ptr, int level, const char* fmt, va_list vl);
void openh264_log_callback(void* ctx, int level, const char* str);

#ifdef __cplusplus
}
#endif


#endif /* LIBV_LOG_H_ */
