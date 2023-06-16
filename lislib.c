#include <stdlib.h>
#include <assert.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

#include <libinsane/capi.h>
#include <libinsane/constants.h>
#include <libinsane/error.h>
#include <libinsane/log.h>
#include <libinsane/safebet.h>
#include <libinsane/util.h>

#include "lislib.h"

void set_log_callbacks() {
	lis_set_log_callbacks(&g_log_callbacks);
}

void lis_value_print(enum lis_value_type typ, union lis_value *val) {
	switch (typ)
	{
	case LIS_TYPE_BOOL:
		printf("bool: %i\n", val->boolean);
		break;
	case LIS_TYPE_DOUBLE:
		printf("double: %f\n", val->dbl);
		break;
	case LIS_TYPE_INTEGER:
		printf("int: %i\n", val->integer);
		break;
	case LIS_TYPE_IMAGE_FORMAT:
		printf("image_format: %i\n", val->format);
		break;
	case LIS_TYPE_STRING:
		printf("string: %s\n", val->string);
		break;
	default:
		printf("UNKNOWN VALUE TYPE\n");
		break;
	}
}

int lis_array_length(void* arr) {
	int i;	
	for (i = 0 ; ((void**)arr)[i] != NULL ; i++) {}
	return i;
}

void lis_value_array_print(enum lis_value_type typ, union lis_value *arr, int count) {
	printf("Array address: %p\n, count: %d", arr, count);
	for (size_t i = 0; i < count; i++)
	{
		printf("elem %p", &arr[i]);
		lis_value_print(typ, &arr[i]);
	}
	
}

void lis_device_array_print(struct lis_device_descriptor **arr) {
	
	for (size_t i = 0; i < lis_array_length((void**)arr); i++)
	{
		printf("device %s\n", arr[i]->dev_id);
		
	}
	
}

union lis_value* lis_option_descriptor_get_value_proxy(struct lis_option_descriptor *opt, struct error_proxy *error) {
	enum lis_error err;
	union lis_value *val;
	val = calloc(1, sizeof(union lis_value)); //won't free it as stated in the get_value's doc
	err = opt->fn.get_value(opt, val);
	if (err != LIS_OK) {
		error->err = err;
		sprintf(error->buf, "Error %d in 'lis_option_descriptor_get_value_proxy' %s", err, lis_strerror(err));
		free(val);
		return NULL;
	}
	return val;
}

void lis_set_option_proxy(struct lis_item *source, char* opt_name, char *value, struct error_proxy *err) {
	union lis_value *val;
	//val = calloc(1, sizeof(union lis_value)); //won't free it as stated in the get_value's doc
	err->err = lis_set_option(source, opt_name, value);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'lis_set_option_proxy' %s", err->err, lis_strerror(err->err));
	}
}

struct lis_scan_session* lis_item_scan_start_proxy(struct lis_item *source, struct error_proxy *err) {
	struct lis_scan_session* session;
	//enum lis_error err;
	err->err = source->scan_start(source, &session);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'lis_item_scan_start_proxy' %s", err->err, lis_strerror(err->err));
		return NULL;
	}
	return session;
}

void lis_scan_session_get_scan_parameters_proxy(struct lis_scan_session *session, struct lis_scan_parameters *params, struct error_proxy *error) {
	enum lis_error err;
	err = session->get_scan_parameters(session, params);
	if (err != LIS_OK) {
		error->err = err;
		sprintf(error->buf, "Error %d in 'lis_scan_session_get_scan_parameters' %s", err, lis_strerror(err));
		return;
	}
}

int lis_scan_session_end_of_feed_proxy(struct lis_scan_session *session) {
	return session->end_of_feed(session);
}

int lis_scan_session_end_of_page_proxy(struct lis_scan_session *session) {
	return session->end_of_page(session);
}

void lis_scan_session_cancel_proxy(struct lis_scan_session *session) {
    return session->cancel(session);
}

void lis_scan_session_scan_read_proxy(struct lis_scan_session *session, void *out_buffer, size_t *buf_size, struct error_proxy *error) {
	enum lis_error err;
	while (1) {
		err = session->scan_read(session, out_buffer, buf_size); 
		if (err == LIS_WARMING_UP) {
			// old scanners need warming time.
			// No data has been returned.
			assert(buf_size == 0);
			logProxy(LIS_LOG_LVL_WARNING, "Warming the lamp up... waiting for 1 sec...");
			sleep(1);
			continue;
		} else {
			break;
		}
	}
	if (err != LIS_OK) {
		error->err = err;
		sprintf(error->buf, "Error %d in 'lis_scan_session_scan_read_proxy' %s", err, lis_strerror(err));
		return;
	}

}

int iterSourcesProxy(void*, struct lis_item*, const char*, enum lis_item_type);

int iterOptionsProxy(void* cb, struct lis_option_descriptor* opt, enum lis_value_type valtype, int conType, void *possible);

void lis_item_iterate_options(struct lis_item *item, void *cb, struct error_proxy *err) {

	//enum lis_error err;
	struct lis_option_descriptor **options;
	struct lis_option_descriptor *opt;
	int i;

	err->err = item->get_options(item, &options);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'iterate_options' %s", err->err, lis_strerror(err->err));		
		return;
	}


	for (i = 0 ; options[i] != NULL ; i++) {	
		opt = options[i];
		if ( !iterOptionsProxy(cb,
								opt,
								opt->value.type,
								opt->constraint.type, 
								&(opt->constraint.possible)) ) {
			break;
		}
		//printf("Option:%s\n", options[i]->name);		
	}	
	item->close(item);
}

struct lis_item* lis_api_get_device_proxy(struct lis_api *api, const char *device_id, struct error_proxy *err) {
	struct lis_item *item;

	err->err = api->get_device(api, device_id, &item);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'lis_api_get_device_proxy' %s", err->err, lis_strerror(err->err));
		return NULL;
	}

	return item;
}

void lis_item_close_proxy(struct lis_item *device) {
	device->close(device);
}

void lis_item_iterate_sources(struct lis_api *api, char *device_id, struct lis_item *device,  void *cb, struct error_proxy *err) {
	struct lis_item *item = device;
	//enum lis_error err;
	struct lis_item **sources;
	int i;

	if (device == NULL) {
		err->err = api->get_device(api, device_id, &item);
		if (err->err != LIS_OK) {
			sprintf(err->buf, "Error %d in 'iterate_sources' %s", err->err, lis_strerror(err->err));
			return;
		}
	}
	//fprintf(stdout, "%s", err == LIS_OK);
	
	err->err = item->get_children(item, &sources);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'iterate_sources' %s", err->err, lis_strerror(err->err));
		return;
	}	

	for (i = 0 ; sources[i] != NULL ; i++) {	
		if ( !iterSourcesProxy(cb, sources[i], sources[i]->name, sources[i]->type) ) {
			break;
		}
	}	

	if (device == NULL) {
		item->close(item);
	}
}


struct lis_item** lis_item_list_sources(struct lis_api *api, char* device_id, struct error_proxy *err) {
	struct lis_item *item = NULL;
	struct lis_item **sources;
	int i;
	err->err = api->get_device(api, device_id, &item);

	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'list_sources' %s", err->err, lis_strerror(err->err));
		return NULL;
	}
	err->err = item->get_children(item, &sources);

	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'list_sources' %s", err->err, lis_strerror(err->err));
		item->close(item);
		return NULL;
	}

	item->close(item);
	return sources;
}

struct lis_api *lis_api_get_api(struct error_proxy *err) {
	struct lis_api *impl = NULL;
	//disable libinsane normalizer to enable BW & Gray scanning
	putenv("LIBINSANE_NORMALIZER_BMP2RAW=0");
	//putenv("LIBINSANE_WORKAROUND_CHECK_CAPABILITIES=0");
	//disable safe default
	//putenv("LIBINSANE_NORMALIZER_SAFE_DEFAULTS=0");

	err->err = lis_safebet(&impl);
	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'get_api' %s", err->err, lis_strerror(err->err));
		return NULL;
	}	
	return impl;
}

void lis_api_cleanup_proxy(struct lis_api* impl) {
	impl->cleanup(impl);
}

struct lis_device_descriptor** lis_api_list_devices_proxy(struct lis_api* impl, struct error_proxy* err) {

	struct lis_device_descriptor **dev_infos;
	struct lis_item *device = NULL;

	err->err = impl->list_devices(impl, LIS_DEVICE_LOCATIONS_ANY, &dev_infos);

	if (err->err != LIS_OK) {
		sprintf(err->buf, "Error %d in 'lis_api_list_devices_proxy' %s", err->err, lis_strerror(err->err));
		return NULL;
	}

	if (dev_infos[0] == NULL) {
		return NULL;
	}
	//lis_device_array_print(dev_infos);
	return dev_infos;
}
