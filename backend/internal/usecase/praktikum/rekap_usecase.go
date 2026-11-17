package usecase

import (
	"fmt"
	"math"
	"strings"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/praktikum"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/praktikum"
)

type RekapUsecase struct {
	rekap repository.RekapRepository
	kelas repository.KelasRepository
}

func NewRekapUsecase(r repository.RekapRepository, k repository.KelasRepository) *RekapUsecase {
	return &RekapUsecase{rekap: r, kelas: k}
}

func (uc *RekapUsecase) GetRekapKelas(kelasID int) (*dto.RekapKelasResponse, error) {
	if _, err := uc.kelas.FindByID(kelasID); err != nil {
		return nil, ErrNotFound
	}

	rows, err := uc.rekap.GetRekapByKelas(kelasID)
	if err != nil {
		return nil, err
	}

	// Query sudah ORDER BY s.urutan, c.id, jadi urutan kolom = urutan kemunculan pertama;
	// map menjaga keunikan, slice menjaga urutan.
	type colMeta struct {
		key   string
		label string
	}
	var columns []colMeta
	colSeen := make(map[string]bool)

	type mhsData struct {
		NIM        string
		Nama       string
		Scores     map[string]dto.RekapSel
		TotalScore float64
	}
	mhsMap := make(map[string]*mhsData)
	var nimOrder []string

	for _, row := range rows {
		if _, ok := mhsMap[row.NIM]; !ok {
			mhsMap[row.NIM] = &mhsData{
				NIM:        row.NIM,
				Nama:       row.Nama,
				Scores:     make(map[string]dto.RekapSel),
				TotalScore: 0,
			}
			nimOrder = append(nimOrder, row.NIM)
		}

		mhs := mhsMap[row.NIM]

		// Jika ada pengerjaan course (sertakan sel walau total_nilai null, asal ada pengerjaan)
		if row.CourseID != nil && row.CourseJenis != nil && row.SesiJudul != nil {
			key := fmt.Sprintf("course_%d", *row.CourseID)

			// Buat label, misalnya "Modul 1 - pretest" (ambil teks sebelum "-").
			shortSesi := *row.SesiJudul
			if parts := strings.SplitN(*row.SesiJudul, "-", 2); len(parts) > 1 {
				shortSesi = strings.TrimSpace(parts[0])
			}
			label := fmt.Sprintf("%s - %s", shortSesi, *row.CourseJenis)

			if !colSeen[key] {
				colSeen[key] = true
				columns = append(columns, colMeta{key: key, label: label})
			}

			isTest := *row.CourseJenis == "pretest" || *row.CourseJenis == "posttest"
			murni := 0.0
			if row.TotalNilai != nil {
				murni = *row.TotalNilai
			}
			sel := dto.RekapSel{Murni: murni, EditKeaktifan: isTest, Total: murni}
			if row.PengerjaanID != nil {
				sel.PengerjaanID = *row.PengerjaanID
			}
			if isTest && row.Keaktifan != nil {
				sel.Keaktifan = row.Keaktifan
				sel.Total = murni + *row.Keaktifan
			}
			// Ujian praktik: konversi 0-100 = murni / max-poin * 100 (dibulatkan).
			if *row.CourseJenis == "ujian_praktik" && row.MaxNilai != nil && *row.MaxNilai > 0 {
				na := math.Round(murni / *row.MaxNilai * 100)
				sel.NilaiAkhir = &na
			}
			mhs.Scores[key] = sel
			mhs.TotalScore += sel.Total
		}
	}

	resp := &dto.RekapKelasResponse{
		Columns: make([]dto.RekapKolom, 0, len(columns)),
		Data:    make([]dto.RekapMahasiswa, 0, len(nimOrder)),
	}

	for _, c := range columns {
		resp.Columns = append(resp.Columns, dto.RekapKolom{
			Key:   c.key,
			Label: c.label,
		})
	}

	for _, nim := range nimOrder {
		md := mhsMap[nim]
		resp.Data = append(resp.Data, dto.RekapMahasiswa{
			NIM:        md.NIM,
			Nama:       md.Nama,
			Scores:     md.Scores,
			TotalScore: md.TotalScore,
		})
	}

	return resp, nil
}
