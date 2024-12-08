package user_usecase

import (
	"context"
	"github.com/DanKo-code/Fitness-Center-Abonement/internal/dtos"
	customErrors "github.com/DanKo-code/Fitness-Center-Abonement/internal/errors"
	"github.com/DanKo-code/Fitness-Center-Abonement/internal/models"
	"github.com/DanKo-code/Fitness-Center-Abonement/internal/repository"
	"github.com/google/uuid"
	"time"
)

type AbonementUseCase struct {
	abonementRepo repository.AbonementRepository
}

func NewAbonementUseCase(abonementRepo repository.AbonementRepository) *AbonementUseCase {
	return &AbonementUseCase{abonementRepo: abonementRepo}
}

func (c *AbonementUseCase) CreateAbonement(
	ctx context.Context,
	cmd *dtos.CreateAbonementCommand,
) (*models.Abonement, error) {

	abonement := &models.Abonement{
		Id:           cmd.Id,
		Title:        cmd.Title,
		Validity:     cmd.Validity,
		VisitingTime: cmd.VisitingTime,
		Photo:        cmd.Photo,
		Price:        cmd.Price,
		UpdatedTime:  time.Now(),
		CreatedTime:  time.Now(),
	}

	createdAbonement, err := c.abonementRepo.CreateAbonement(ctx, abonement)
	if err != nil {
		return nil, err
	}

	return createdAbonement, nil
}

func (c *AbonementUseCase) GetAbonementById(
	ctx context.Context,
	id uuid.UUID,
) (*models.Abonement, error) {
	abonement, err := c.abonementRepo.GetAbonementById(ctx, id)
	if err != nil {
		return nil, err
	}

	return abonement, nil
}

func (c *AbonementUseCase) UpdateAbonement(
	ctx context.Context,
	cmd *dtos.UpdateAbonementCommand,
) (*models.Abonement, error) {

	err := c.abonementRepo.UpdateAbonement(ctx, cmd)
	if err != nil {
		return nil, err
	}

	abonement, err := c.abonementRepo.GetAbonementById(ctx, cmd.Id)
	if err != nil {
		return nil, err
	}

	return abonement, nil
}

func (c *AbonementUseCase) DeleteAbonementById(
	ctx context.Context,
	id uuid.UUID,
) (*models.Abonement, error) {
	abonement, err := c.abonementRepo.GetAbonementById(ctx, id)
	if err != nil {
		return nil, customErrors.AbonementNotFound
	}

	err = c.abonementRepo.DeleteAbonementById(ctx, id)
	if err != nil {
		return nil, err
	}

	return abonement, nil
}

func (c *AbonementUseCase) GetAbonementes(
	ctx context.Context,
) ([]*models.Abonement, error) {

	abonementes, err := c.abonementRepo.GetAbonementes(ctx)
	if err != nil {
		return nil, err
	}

	return abonementes, nil
}
