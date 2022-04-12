package internals

func memcpy(dest []byte, src []byte, size uint) {
	var i uint
	for i = 0; i < size; i++ {
		dest[i] = src[i]
	}
}
