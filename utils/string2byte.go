package utils

import "unsafe"


/**
	字符串 和 字节数组的 高效转化  （避免拷贝）

	性能优化：
		字符串 转成 字节数组(slice, []byte) 时会对底层字节数组的复制

	字符串的底层实现：
			uint8 *str
			int len

	字节数组的底层实现：
			uint8 *array
			int len
			int cap

	当字符串直接强制类型转换到字节数组时，会发生数据拷贝，影响性能


	高效处理思路：
		[2]uintptr    ---->    [3]uintptr
 */


func Str2bytes(s string) []byte {
	// [2]uintptr    --->  [3]uintptr
	sp := (*[2]uintptr)(unsafe.Pointer(&s))

	bp := [3]uintptr{sp[0], sp[1], sp[1]}

	ret_p := (*[]byte)(unsafe.Pointer(&bp))  // 强转成 []byte指针
	return *ret_p
}


func Bytes2str(b []byte) string {
	sp := (*string)(unsafe.Pointer(&b))     // 强转成 string 指针

	return *sp
}

