package checksum

type NullWriter struct{}

func NewNullWriter() *NullWriter {
	return &NullWriter{}
}

func (w *NullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
