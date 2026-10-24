package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

// Prompter задаёт вопрос и переспрашивает, пока не получит нормальный ответ
type prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func (p prompter) line(question string) (string, error) {
	fmt.Fprintf(p.out, "%s: ", question)
	s, err := p.in.ReadString('\n')
	if errors.Is(err, io.EOF) && s != "" {
		// Последняя строка без перевода строки, это нормальный ответ.
		err = nil
	}
	if err != nil {
		if errors.Is(err, io.EOF) {
			fmt.Fprintln(p.out)
		}
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func (p prompter) text(question string) (string, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return "", err
		}
		if s != "" {
			return s, nil
		}
		fmt.Fprintln(p.out, "  Значение не может быть пустым.")
	}
}

func (p prompter) intInRange(question string, lo, hi int) (int, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return 0, err
		}
		n, convErr := strconv.Atoi(s)
		if convErr == nil && n >= lo && n <= hi {
			return n, nil
		}
		fmt.Fprintf(p.out, "  Введите целое число от %d до %d.\n", lo, hi)
	}
}

// floatInRange понимает числа с плавующей точкой через точку и запятую (0.5, и 0,5)
func (p prompter) floatInRange(question string, lo, hi float64) (float64, error) {
	for {
		s, err := p.line(question)
		if err != nil {
			return 0, err
		}
		x, convErr := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
		if convErr == nil && !math.IsNaN(x) && x >= lo && x <= hi {
			return x, nil
		}
		fmt.Fprintf(p.out, "  Введите число от %s до %s, например 0,5.\n", formatNum(lo), formatNum(hi))
	}
}

func (p prompter) yesNo(question string) (bool, error) {
	for {
		s, err := p.line(question + " (д/н)")
		if err != nil {
			return false, err
		}
		switch strings.ToLower(s) {
		case "д", "да", "y", "yes":
			return true, nil
		case "н", "нет", "n", "no":
			return false, nil
		}
		fmt.Fprintln(p.out, "  Ответьте «д» или «н».")
	}
}

func formatNum(x float64) string {
	return strings.Replace(strconv.FormatFloat(x, 'f', -1, 64), ".", ",", 1)
}
