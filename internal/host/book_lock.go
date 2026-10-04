package host

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
)

const bookLockFile = ".ainovel.lock"

// ErrBookInUse 表示同一小说目录已被另一个进程占用。
var ErrBookInUse = errors.New("thư mục tiểu thuyết đã bị chiếm dụng bởi một tiến trình ainovel-cli khác")

// bookLease 在 Host 的完整生命周期内持有小说目录的跨进程独占权。
// 锁文件会保留在目录中；真正的占用状态由操作系统管理，进程异常退出也会自动释放。
type bookLease struct {
	lock *flock.Flock
}

func isAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	if os.IsPermission(err) || errors.Is(err, os.ErrPermission) {
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "access is denied") || strings.Contains(s, "permission denied")
}

func acquireBookLease(dir string) (*bookLease, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("phân tích thư mục tiểu thuyết: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		if isAccessDenied(err) {
			return nil, fmt.Errorf("lỗi quyền truy cập: không thể tạo thư mục tiểu thuyết %q (Access is denied). Vui lòng cấp quyền ghi hoặc sử dụng --dir chỉ định thư mục khác: %w", absDir, err)
		}
		return nil, fmt.Errorf("tạo thư mục tiểu thuyết: %w", err)
	}
	fileLock := flock.New(filepath.Join(absDir, bookLockFile), flock.SetPermissions(0o666))
	locked, err := fileLock.TryLock()
	if err != nil {
		if isAccessDenied(err) {
			return nil, closeBookLockAfterFailure(fileLock, fmt.Errorf(
				"chiếm dụng thư mục tiểu thuyết %q: không thể tạo hoặc mở tệp khóa %s (Access is denied). Vui lòng kiểm tra quyền ghi thư mục hoặc đóng tiến trình terminal khác đang giữ tệp này: %w",
				absDir, bookLockFile, err,
			))
		}
		return nil, closeBookLockAfterFailure(fileLock, fmt.Errorf("chiếm dụng thư mục tiểu thuyết %q: %w", absDir, err))
	}
	if !locked {
		return nil, closeBookLockAfterFailure(fileLock, fmt.Errorf(
			"%w: %s; vui lòng đóng terminal khác đang thao tác thư mục này, hoặc sử dụng thư mục tiểu thuyết khác",
			ErrBookInUse,
			absDir,
		))
	}
	return &bookLease{lock: fileLock}, nil
}

func closeBookLockAfterFailure(fileLock *flock.Flock, cause error) error {
	if err := fileLock.Close(); err != nil {
		return errors.Join(cause, fmt.Errorf("đóng khóa thư mục tiểu thuyết: %w", err))
	}
	return cause
}

func (l *bookLease) Close() error {
	if l == nil || l.lock == nil {
		return nil
	}
	err := l.lock.Close()
	l.lock = nil
	return err
}
