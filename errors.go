package phash

// DecodeError describes failures in HTTP setup, HTTP status, IO reads, or image decoding.
// Returned by the helpers in decode.go to avoid raw fmt.Errorf strings.
type DecodeOp string

const (
	DecodeOpRequest    DecodeOp = "request"
	DecodeOpHTTP       DecodeOp = "http"
	DecodeOpHTTPStatus DecodeOp = "http status"
	DecodeOpRead       DecodeOp = "read"
	DecodeOpDecode     DecodeOp = "decode"
)

type DecodeError struct {
	Op  DecodeOp
	Err error
}

// Error formats DecodeError as "op: err" (or "op" when Err is nil).
func (e DecodeError) Error() string {
	if e.Err == nil {
		return string(e.Op)
	}
	return string(e.Op) + ": " + e.Err.Error()
}

// EncodeError describes failures when encoding images.
// Returned by helpers in encode.go to avoid raw fmt.Errorf strings.
type EncodeOp string

const (
	EncodeOpWebP EncodeOp = "webp encode"
)

type EncodeError struct {
	Op  EncodeOp
	Err error
}

// Error formats EncodeError as "op: err" (or "op" when Err is nil).
func (e EncodeError) Error() string {
	if e.Err == nil {
		return string(e.Op)
	}
	return string(e.Op) + ": " + e.Err.Error()
}
