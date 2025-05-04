package random

/**
Go標準ライブラリの乱数生成器をラップして、スレッドセーフ(複数のゴルーチンから同時にアクセスしても安全)にするためのラッパーを実装する
*/
import (
	"math/rand/v2"
	"sync"
	"time" // time パッケージをインポート
)

type lockedSource struct {
	inner rand.Source
	sync.Mutex
}

func (r *lockedSource) Uint64() uint64 {
	r.Lock()
	defer r.Unlock()
	return r.inner.Uint64()
}

func NewLockedSource(src rand.Source) rand.Source {
	return &lockedSource{
		inner: src,
	}
}

func CreateChildSource(parent rand.Source) rand.Source {
	return rand.NewPCG(parent.Uint64(), parent.Uint64())
}

func CreateChildRand(parent rand.Source) *rand.Rand {
	return rand.New(NewLockedSource(CreateChildSource(parent)))
}

// GenerateUserName generates a random user name.
// TODO: Implement actual random generation logic.
func GenerateUserName() string {
	// Note: This is a placeholder implementation.
	// Replace with actual random generation logic later.
	return "placeholder_user"
}

// GenerateFirstName generates a random first name.
// TODO: Implement actual random generation logic.
func GenerateFirstName() string {
	// Note: This is a placeholder implementation.
	// Replace with actual random generation logic later.
	return "Placeholder"
}

// GenerateLastName generates a random last name.
// TODO: Implement actual random generation logic.
func GenerateLastName() string {
	// Note: This is a placeholder implementation.
	// Replace with actual random generation logic later.
	return "User"
}

// GenerateDateOfBirth generates a random date of birth.
// TODO: Implement actual random generation logic.
func GenerateDateOfBirth() string {
	// Note: This is a placeholder implementation.
	// Replace with actual random generation logic later.
	// time パッケージを使って日付文字列を生成
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}
