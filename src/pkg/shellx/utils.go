package shellx

import "bytes"

type streamWriter struct {
	buf    *bytes.Buffer
	onData func(string)
}

func (w *streamWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	if err != nil {
		return n, err
	}
	if w.onData != nil {
		w.onData(string(p))
	}
	return n, nil
}
