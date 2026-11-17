package usecase

import (
	"context"

	dto "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/dto/elearning"
	repository "github.com/Lab-Algoritma-dan-Pemrograman/algohub/backend/internal/repository/elearning"
)

type PlaygroundUsecase struct {
	examples repository.PlaygroundRepository
}

func NewPlaygroundUsecase(examples repository.PlaygroundRepository) *PlaygroundUsecase {
	return &PlaygroundUsecase{examples: examples}
}

func (uc *PlaygroundUsecase) List(ctx context.Context, p *dto.Principal, language string, page, limit int) (*dto.Page[dto.PlaygroundExample], error) {
	return listPage(p, page, limit, func() ([]dto.PlaygroundExample, int64, error) {
		rows, total, err := uc.examples.List(ctx, language, page, limit)
		out := make([]dto.PlaygroundExample, len(rows))
		for i, e := range rows {
			out[i] = dto.PlaygroundExample{ID: e.ID, Title: e.Title, Language: e.Language, Code: e.Code, Description: e.Description, SortOrder: e.SortOrder}
		}
		return out, total, err
	})
}

// Kolom persis addPlaygroundExample; sortOrder yang tidak dikirim memakai default DB seperti kunci undefined di JSON lama.
func (uc *PlaygroundUsecase) Add(ctx context.Context, p *dto.Principal, req *dto.PlaygroundInput) (*dto.IDResponse, error) {
	if err := requirePrincipal(p); err != nil {
		return nil, err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil, RLSInsert("playground_examples")
	}
	cols := map[string]any{"title": req.Title, "language": req.Language, "code": req.Code, "description": req.Description}
	if req.SortOrder != nil {
		cols["sort_order"] = *req.SortOrder
	}
	id, err := uc.examples.Insert(ctx, cols)
	if err != nil {
		return nil, err
	}
	return &dto.IDResponse{ID: id}, nil
}

func (uc *PlaygroundUsecase) Update(ctx context.Context, p *dto.Principal, id string, req *dto.PlaygroundUpdate) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	cols := filledColumns(map[string]string{"title": req.Title, "language": req.Language, "code": req.Code})
	if req.Description.Set {
		cols["description"] = req.Description.Value
	}
	if req.SortOrder.Set {
		cols["sort_order"] = req.SortOrder.Value
	}
	_, err := uc.examples.Update(ctx, id, cols)
	return err
}

func (uc *PlaygroundUsecase) Delete(ctx context.Context, p *dto.Principal, id string) error {
	if err := requirePrincipal(p); err != nil {
		return err
	}
	if !IsAdminOrKordas(p.Role) {
		return nil
	}
	_, err := uc.examples.Delete(ctx, id)
	return err
}
