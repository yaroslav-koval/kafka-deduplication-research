package cerror

func IsNotExist(err error) bool {
	kerr, ok := err.(KindError)
	if !ok {
		return false
	}

	kind := kerr.Kind()

	return kind == KindNotExist ||
		kind == KindDBNoRows ||
		kind == KindMinioNotExist
}
