package usecase

import (
	"testing"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	entity "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/entity/praktikum"
	"github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func intPtr(i int) *int       { return &i }
func strPtr(s string) *string { return &s }

// attrsHas membandingkan map attrs persis (kunci dan nilai).
func attrsHas(want map[string]any) interface{} {
	return mock.MatchedBy(func(got map[string]any) bool {
		if len(got) != len(want) {
			return false
		}
		for k, v := range want {
			gv, ok := got[k]
			if !ok || gv != v {
				return false
			}
		}
		return true
	})
}

// PATCH diperlakukan PUT: request parsial TIDAK boleh mengosongkan field
// opsional (kelas_id/shift/gelombang/kelompok). Field yang pointer-nya nil
// tidak boleh muncul di attrs yang dikirim ke repository.
func TestUpdateMahasiswa_PartialRequest_DoesNotClearOptionalFields(t *testing.T) {
	existing := &entity.User{
		ID:        99,
		NIM:       "202411001",
		Nama:      "LAMA",
		KelasID:   intPtr(3),
		Shift:     intPtr(1),
		Gelombang: intPtr(2),
		Kelompok:  strPtr("A"),
	}

	users := mocks.NewUserRepository(t)
	// UpdateMahasiswa memanggil FindByID di awal dan di akhir (return).
	users.On("FindByID", 99).Return(existing, nil).Times(2)
	// Hanya nim+nama boleh dikirim.
	users.On("BulkUpdateAttrs", []string{"202411001"}, attrsHas(map[string]any{
		"nim":  "202411001",
		"nama": "BARU",
	})).Return(int64(1), nil).Once()

	uc := NewUserUsecase(users, mocks.NewKelasRepository(t))

	req := &dto.UserRequest{NIM: "202411001", Nama: "BARU"} // sisanya nil
	_, err := uc.UpdateMahasiswa(99, req)
	assert.NoError(t, err)
	users.AssertExpectations(t)
}

// Field opsional yang dikirim harus ikut masuk attrs.
func TestUpdateMahasiswa_OptionalFieldsForwarded(t *testing.T) {
	existing := &entity.User{ID: 5, NIM: "202411005", Nama: "X"}

	users := mocks.NewUserRepository(t)
	users.On("FindByID", 5).Return(existing, nil).Times(2)
	users.On("BulkUpdateAttrs", []string{"202411005"}, attrsHas(map[string]any{
		"nim":       "202411005",
		"nama":      "X",
		"kelas_id":  7,
		"shift":     2,
		"gelombang": 1,
		"kelompok":  "C",
	})).Return(int64(1), nil).Once()

	uc := NewUserUsecase(users, mocks.NewKelasRepository(t))
	req := &dto.UserRequest{NIM: "202411005", Nama: "X", KelasID: intPtr(7), Shift: intPtr(2), Gelombang: intPtr(1), Kelompok: strPtr("C")}
	_, err := uc.UpdateMahasiswa(5, req)
	assert.NoError(t, err)
	users.AssertExpectations(t)
}

// Negative check: kalau kode lama (full-struct Save) dipakai, attrs yang
// dikirim akan berisi kunci opsional walau pointer nil -> test di atas FAIL.
// Diverifikasi dengan membalik patch sementara; tidak di-commit.