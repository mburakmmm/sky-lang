package skylib

import (
	"bufio"
	"io"
	"os"
)

// IOReader wraps an io.Reader
type IOReader struct {
	reader io.Reader
}

// IOWriter wraps an io.Writer
type IOWriter struct {
	writer io.Writer
}

// NewReader creates a new reader
func IONewReader(r io.Reader) *IOReader {
	return &IOReader{reader: r}
}

// NewWriter creates a new writer
func IONewWriter(w io.Writer) *IOWriter {
	return &IOWriter{writer: w}
}

// Read reads up to n bytes
func (r *IOReader) Read(n int) ([]byte, error) {
	buf := make([]byte, n)
	bytesRead, err := r.reader.Read(buf)
	return buf[:bytesRead], err
}

// ReadAll reads all bytes
func (r *IOReader) ReadAll() ([]byte, error) {
	return io.ReadAll(r.reader)
}

// Write writes bytes
func (w *IOWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

// WriteString writes string
func (w *IOWriter) WriteString(s string) (int, error) {
	return w.writer.Write([]byte(s))
}

// Stdin returns standard input
func IOStdin() *IOReader {
	return &IOReader{reader: os.Stdin}
}

// Stdout returns standard output
func IOStdout() *IOWriter {
	return &IOWriter{writer: os.Stdout}
}

// Stderr returns standard error
func IOStderr() *IOWriter {
	return &IOWriter{writer: os.Stderr}
}

// ReadLine reads a line from stdin
func IOReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line[:len(line)-1], nil // Remove newline
}

// Copy copies from reader to writer
func IOCopy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

// BufReader creates buffered reader
type BufReader struct {
	*bufio.Reader
}

// NewBufReader creates a buffered reader
func IONewBufReader(r io.Reader) *BufReader {
	return &BufReader{bufio.NewReader(r)}
}

// ReadLine reads a line
func (br *BufReader) ReadLine() (string, error) {
	return br.ReadString('\n')
}

// BufWriter creates buffered writer
type BufWriter struct {
	*bufio.Writer
}

// NewBufWriter creates a buffered writer
func IONewBufWriter(w io.Writer) *BufWriter {
	return &BufWriter{bufio.NewWriter(w)}
}

// Flush flushes the buffer
func (bw *BufWriter) Flush() error {
	return bw.Writer.Flush()
}

