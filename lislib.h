#ifndef __LISLIB_H
#define __LISLIB_H

#include <libinsane/log.h>
#include <libinsane/error.h>
#include <libinsane/util.h>


struct error_proxy {
	char *buf;
	enum lis_error err;
};

//lis_api functions
struct lis_device_descriptor**  lis_api_list_devices_proxy(struct lis_api*, struct error_proxy*);
struct lis_item*                lis_api_get_device_proxy(struct lis_api*, const char*, struct error_proxy*);
struct lis_api*                 lis_api_get_api(struct error_proxy*);
void                            lis_api_cleanup_proxy(struct lis_api*);

//lis_item functions
void                        lis_item_close_proxy(struct lis_item*);
struct lis_scan_session*    lis_item_scan_start_proxy(struct lis_item*, struct error_proxy*);
void                        lis_item_iterate_sources(struct lis_api*, char*, struct lis_item*,  void*, struct error_proxy*);
struct lis_item**           lis_item_list_sources(struct lis_api*, char*, struct error_proxy*);
void                        lis_item_iterate_options(struct lis_item*, void*, struct error_proxy*);

//lis_scan_session functions
void    lis_scan_session_get_scan_parameters_proxy(struct lis_scan_session*, struct lis_scan_parameters*, struct error_proxy*);
int     lis_scan_session_end_of_feed_proxy(struct lis_scan_session*);
int     lis_scan_session_end_of_page_proxy(struct lis_scan_session*);
void    lis_scan_session_cancel_proxy(struct lis_scan_session*);
void    lis_scan_session_scan_read_proxy(struct lis_scan_session*, void*, size_t*, struct error_proxy*);

//lis_value debugging functions
void lis_value_print(enum lis_value_type, union lis_value*);
void lis_value_array_print(enum lis_value_type, union lis_value*, int);

//lis_option functions
union lis_value* lis_option_descriptor_get_value_proxy(struct lis_option_descriptor*, struct error_proxy*);
void             lis_set_option_proxy(struct lis_item*, char*, char*, struct error_proxy*);

//utils functions
int  lis_array_length(void*);
void logProxy(enum lis_log_level, char*);

//various debug functions
void go_printf(const char*);
void set_log_callbacks();


//this is to silence gcc complains about const <-> no const function parameter conversions
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types"
static struct lis_log_callbacks g_log_callbacks = {
	.callbacks = {
		[LIS_LOG_LVL_DEBUG] = logProxy,
		[LIS_LOG_LVL_INFO] = logProxy,
		[LIS_LOG_LVL_WARNING] = logProxy,
		[LIS_LOG_LVL_ERROR] = logProxy,
	},
};
#pragma GCC diagnostic warning "-Wincompatible-pointer-types"


#endif