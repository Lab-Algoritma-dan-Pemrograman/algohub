package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/elearning"
	usecase "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/usecase/elearning"
)

func TestActivityLogUsesAuthNIMAndBodyName(t *testing.T) {
	f := newLearn(t)
	p := mismatched(usecase.RolePraktikan)
	f.logs.On("Insert", mock.Anything, mock.MatchedBy(func(l *entity.ActivityLog) bool {
		return *l.NIM == "sub-lain" && *l.Nama == "Budi" && l.EventType == "logout" && *l.Details == "Melakukan logout dari sistem"
	})).Return(nil)
	err := f.logUC.Insert(bg, p, &dto.ActivityLogRequest{EventType: "logout", Details: "Melakukan logout dari sistem", Nama: str("Budi")})
	require.NoError(t, err)
}

func TestActivityLogRequiresPrincipal(t *testing.T) {
	err := newLearn(t).logUC.Insert(bg, nil, &dto.ActivityLogRequest{EventType: "login"})
	assertMissingToken(t, err)
}
