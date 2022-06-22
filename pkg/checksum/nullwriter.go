package checksum

type NullWriter struct{}

func (w *NullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
