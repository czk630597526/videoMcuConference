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

//全局日志函数
#define input_error                   libv_log(E_LIBV_LOG_ERROR, __FILE__, __LINE__, "The funtion input is error!");
#define elog(...)                   libv_log(E_LIBV_LOG_ERROR, __FILE__, __LINE__, __VA_ARGS__)
#define wlog(...)                   libv_log(E_LIBV_LOG_WARING, __FILE__, __LINE__, __VA_ARGS__)
#define ilog(...)                   libv_log(E_LIBV_LOG_INFO, __FILE__, __LINE__, __VA_ARGS__)
#define dlog(...)                   libv_log(E_LIBV_LOG_DEBUG, __FILE__, __LINE__, __VA_ARGS__)

typedef void ( * LIBV_LOG_FUN )(LIBV_LOG_LEVEL_E level, const char* file_name, int line_no, const char* format, va_list vl);

void libv_log(LIBV_LOG_LEVEL_E level, const char* file_name, int line_no, const char* format, ...);

/*
 如果log_cbfun不为空，则回调上层的日志函数。否则，写入到文件log_file中;printf_flg:是否打印至前台
 如果是多线程调用，则要上层回调处理日志，底层不支持多线程处理日志。
 * */
void libv_log_set(LIBV_LOG_FUN log_cbfun, LIBV_LOG_LEVEL_E level, const char* log_file, int printf_flg);

void ffm_av_log_callback(void* ptr, int level, const char* fmt, va_list vl);
void openh264_log_callback(void* ctx, int level, const char* str);

#ifdef __cplusplus
}
#endif


#endif /* LIBV_LOG_H_ */
