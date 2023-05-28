package main

import "C"

/*
const char* build_time(void)
{
	static const char* psz_build_time = "["__DATE__ " " __TIME__ "]";
	return psz_build_time;
}
*/
import "C"

const version = "Mcu_Conference_Simple"

func BuildTime() string {
	return C.GoString(C.build_time()) + "  |  " + version
}
