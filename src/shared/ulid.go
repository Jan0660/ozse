package shared

import (
	"github.com/oklog/ulid/v2"
	"math/rand"
	"time"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func NewUlid() ulid.ULID {
	return ulid.MustNew(ulid.Now(), entropy)
}
