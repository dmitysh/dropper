package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Fatal(_ context.Context, msg ...any) {
	fmt.Println(append([]any{color.RedString("Fatal error:")}, msg...)...)
	os.Exit(1)
}

func FatalIfErr(ctx context.Context, err error, msg ...any) {
	if err != nil {
		Fatal(ctx, []any{err, msg})
	}
}

func Err(_ context.Context, msg ...any) {
	fmt.Println(append([]any{color.RedString("Error:")}, msg...)...)
}

func Warn(_ context.Context, msg ...any) {
	fmt.Println(append([]any{color.YellowString("Warning:")}, msg...)...)
}

func Info(_ context.Context, msg ...any) {
	fmt.Println(append([]any{color.GreenString("Info:")}, msg...)...)
}

func Fatalf(_ context.Context, f string, v ...any) {
	fmt.Printf(color.RedString("\nFatal error: ")+f+"\n", v...)
	os.Exit(1)
}

func FatalfIfErr(ctx context.Context, err error, format string, a ...any) {
	if err != nil {
		Fatalf(ctx, format, a...)
	}
}

func Errf(ctx context.Context, f string, v ...any) {
	fmt.Printf(color.RedString("Error: ")+f+"\n", v...)
}

func Warnf(ctx context.Context, f string, v ...any) {
	fmt.Printf(color.YellowString("Warning: ")+f+"\n", v...)
}

func Infof(ctx context.Context, f string, v ...any) {
	fmt.Printf(color.GreenString("Info: ")+f+"\n", v...)
}
