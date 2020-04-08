package repository

import (
	"context"
	"github.com/caos/zitadel/internal/project/model"
)

type ProjectRepository interface {
	ProjectByID(ctx context.Context, id string) (*model.Project, error)
	CreateProject(ctx context.Context, name string) (*model.Project, error)
	UpdateProject(ctx context.Context, project *model.Project) (*model.Project, error)
	DeactivateProject(ctx context.Context, id string) (*model.Project, error)
	ReactivateProject(ctx context.Context, id string) (*model.Project, error)
}