package fancylog

import "go.uber.org/zap/buffer"

type Color []byte

// UncheckedCustomColor wrapper, this does not validate color
func UncheckedCustomColor(b []byte) Color {
	return Color(b)
}

type ColorLogger struct {
	*buffer.Buffer
}

var (
	_pool = buffer.NewPool()
	// Get retrieves a buffer from the pool, creating one if necessary.
	Get = _pool.Get
)

var (
	colorOff    Color = []byte("\u001B[0m")
	ColorRed    Color = []byte("\u001B[0;31m")
	ColorGreen  Color = []byte("\u001B[0;32m")
	ColorOrange Color = []byte("\u001B[0;33m")
	ColorBlue   Color = []byte("\u001B[0;34m")
	ColorPurple Color = []byte("\u001B[0;35m")
	ColorCyan   Color = []byte("\u001B[0;36m")
	ColorGray   Color = []byte("\u001B[0;37m")

	ColorFatalRed    Color = []byte("\u001b[1m\u001b[31m\u001b[7m")
	ColorDarkOrange  Color = []byte("\u001b[1m\u001b[38;5;202m")
	ColorBrightWhite Color = []byte("\u001b[1m\u001b[38;5;255m")
	ColorNicePurple  Color = []byte("\u001b[1m\u001b[38;5;99m")
)

func NewColorLogger() ColorLogger {
	return ColorLogger{Get()}
}

// Off apply no color to the data
func (cb *ColorLogger) Off() {
	_, _ = cb.Write(colorOff)
}

// Red apply red color to the data
func (cb *ColorLogger) Red() {
	_, _ = cb.Write(ColorRed)
}

// Green apply green color to the data
func (cb *ColorLogger) Green() {
	_, _ = cb.Write(ColorGreen)
}

// Orange apply orange color to the data
func (cb *ColorLogger) Orange() {
	_, _ = cb.Write(ColorOrange)
}

// Blue apply blue color to the data
func (cb *ColorLogger) Blue() {
	_, _ = cb.Write(ColorBlue)
}

// Purple apply purple color to the data
func (cb *ColorLogger) Purple() {
	_, _ = cb.Write(ColorPurple)
}

// Cyan apply cyan color to the data
func (cb *ColorLogger) Cyan() {
	_, _ = cb.Write(ColorCyan)
}

// Gray apply gray color to the data
func (cb *ColorLogger) Gray() {
	_, _ = cb.Write(ColorGray)
}

// White apply gray color to the data
func (cb *ColorLogger) White() {
	_, _ = cb.Write(ColorBrightWhite)
}

// BrightOrange apply gray color to the data
func (cb *ColorLogger) BrightOrange() {
	_, _ = cb.Write(ColorDarkOrange)
}

// NicePurple apply gray color to the data
func (cb *ColorLogger) NicePurple() {
	_, _ = cb.Write(ColorNicePurple)
}

// WriteColor apply given color
func (cb *ColorLogger) WriteColor(color Color) {
	_, _ = cb.Write(color)
}

func (cb *ColorLogger) AppendSpace() {
	_, _ = cb.Write([]byte(" "))
}

// Append byte slice to buffer
func (cb *ColorLogger) Append(data []byte) {
	_, _ = cb.Write(data)
}

// AppendWithColor byte slice to buffer
func (cb *ColorLogger) AppendWithColor(data []byte, color Color) {
	_, _ = cb.Write(mixer(data, color))
}

// mixer mix the color on and off byte with the actual data
func mixer(data []byte, color []byte) []byte {
	var result []byte
	return append(append(append(result, color...), data...), colorOff...)
}

// Red apply red color to the data
func Red(data []byte) []byte {
	return mixer(data, ColorRed)
}

// Green apply green color to the data
func Green(data []byte) []byte {
	return mixer(data, ColorGreen)
}

// Orange apply orange color to the data
func Orange(data []byte) []byte {
	return mixer(data, ColorOrange)
}

// Blue apply blue color to the data
func Blue(data []byte) []byte {
	return mixer(data, ColorBlue)
}

// Purple apply purple color to the data
func Purple(data []byte) []byte {
	return mixer(data, ColorPurple)
}

// Cyan apply cyan color to the data
func Cyan(data []byte) []byte {
	return mixer(data, ColorCyan)
}

// Gray apply gray color to the data
func Gray(data []byte) []byte {
	return mixer(data, ColorGray)
}
