package logs

import (
	"log/slog"

	"go.uber.org/zap"
)

func okLogs() {
	logger := zap.NewExample()

	slog.Info("user created", "user_id", 1)
	logger.Info("request succeeded", zap.Int("code", 200))
}

func badLogs() {
	logger := zap.NewExample()

	slog.Info("User created", "user_id", 1)              // want "log message should start with a lowercase letter"
	slog.Info("пользователь создан")                    // want "log message should contain only English letters"
	slog.Info("user created ✅")                        // want "log message should not contain special symbols or emoji"
	slog.Info("user password leaked")                   // want "log message should not contain potentially sensitive data"
	logger.Info("Password is %s", zap.String("x", "y")) // want "log message should start with a lowercase letter" "log message should not contain potentially sensitive data"
}

