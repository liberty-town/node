package globals

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func Test_Oracles(t *testing.T) {

	converted, err := ConvertCurrencyToAsset("DOLLAR", "PCASH", rand.Uint64())
	assert.Nil(t, err, "err")

	fmt.Println(converted)

}
