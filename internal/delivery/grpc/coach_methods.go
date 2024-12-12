package grpc

import (
	"context"
	"errors"
	"github.com/DanKo-code/Fitness-Center-Abonement/internal/dtos"
	customErrors "github.com/DanKo-code/Fitness-Center-Abonement/internal/errors"
	"github.com/DanKo-code/Fitness-Center-Abonement/internal/usecase"
	"github.com/DanKo-code/Fitness-Center-Abonement/pkg/logger"
	abonementProtobuf "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"time"
)

var _ abonementProtobuf.AbonementServer = (*AbonementgRPC)(nil)

type AbonementgRPC struct {
	abonementProtobuf.UnimplementedAbonementServer

	abonementUseCase usecase.AbonementUseCase
	cloudUseCase     usecase.CloudUseCase
}

func RegisterAbonementServer(gRPC *grpc.Server, abonementUseCase usecase.AbonementUseCase, cloudUseCase usecase.CloudUseCase) {
	abonementProtobuf.RegisterAbonementServer(gRPC, &AbonementgRPC{abonementUseCase: abonementUseCase, cloudUseCase: cloudUseCase})
}

func (c *AbonementgRPC) CreateAbonement(g grpc.ClientStreamingServer[abonementProtobuf.CreateAbonementRequest, abonementProtobuf.CreateAbonementResponse]) error {

	abonementData, abonementPhoto, err := GetObjectData(
		&g,
		func(chunk *abonementProtobuf.CreateAbonementRequest) interface{} {
			return chunk.GetAbonementDataForCreate()
		},
		func(chunk *abonementProtobuf.CreateAbonementRequest) []byte {
			return chunk.GetAbonementPhoto()
		},
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request data")
	}

	if abonementData == nil {
		logger.ErrorLogger.Printf("abonement data is empty")
		return status.Error(codes.InvalidArgument, "abonement data is empty")
	}

	castedAbonementData, ok := abonementData.(*abonementProtobuf.AbonementDataForCreate)
	if !ok {
		logger.ErrorLogger.Printf("abonement data is not of type AbonementProtobuf.AbonementDataForCreate")
		return status.Error(codes.InvalidArgument, "abonement data is not of type AbonementProtobuf.AbonementDataForCreate")
	}

	cmd := &dtos.CreateAbonementCommand{
		Id:           uuid.New(),
		Title:        castedAbonementData.Title,
		Validity:     castedAbonementData.Validity,
		VisitingTime: castedAbonementData.VisitingTime,
		Price:        int(castedAbonementData.Price),
	}

	var photoURL string
	if abonementPhoto != nil {
		url, err := c.cloudUseCase.PutObject(context.TODO(), abonementPhoto, "abonement/"+cmd.Id.String())
		photoURL = url
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create abonement photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create abonement photo in cloud")
		}
	}

	cmd.Photo = photoURL

	abonement, err := c.abonementUseCase.CreateAbonement(context.TODO(), cmd)
	if err != nil {

		if photoURL == "" {
			err := c.cloudUseCase.DeleteObject(context.TODO(), "abonement/"+cmd.Id.String())
			if err != nil {
				logger.ErrorLogger.Printf("Failed to delete abonement photo from cloud: %v", err)
				return status.Error(codes.Internal, "Failed to delete abonement photo in cloud")
			}
		}

		return status.Error(codes.Internal, "Failed to create abonement")
	}

	abonementObject := &abonementProtobuf.AbonementObject{
		Id:           abonement.Id.String(),
		Title:        abonement.Title,
		Validity:     abonement.Validity,
		VisitingTime: abonement.VisitingTime,
		Photo:        abonement.Photo,
		Price:        int32(abonement.Price),
		CreatedTime:  abonement.CreatedTime.String(),
		UpdatedTime:  abonement.UpdatedTime.String(),
	}

	response := &abonementProtobuf.CreateAbonementResponse{
		AbonementObject: abonementObject,
	}

	err = g.SendAndClose(response)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to send abonement create response: %v", err)
		return status.Error(codes.Internal, "Failed to send abonement create response")
	}

	return nil
}

func (c *AbonementgRPC) GetAbonementById(ctx context.Context, request *abonementProtobuf.GetAbonementByIdRequest) (*abonementProtobuf.GetAbonementByIdResponse, error) {

	abonement, err := c.abonementUseCase.GetAbonementById(ctx, uuid.MustParse(request.Id))
	if err != nil {

		if errors.Is(err, customErrors.AbonementNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, err
	}

	abonementObject := &abonementProtobuf.AbonementObject{
		Id:           abonement.Id.String(),
		Title:        abonement.Title,
		Validity:     abonement.Validity,
		VisitingTime: abonement.VisitingTime,
		Photo:        abonement.Photo,
		Price:        int32(abonement.Price),
		CreatedTime:  abonement.CreatedTime.String(),
		UpdatedTime:  abonement.UpdatedTime.String(),
	}

	response := &abonementProtobuf.GetAbonementByIdResponse{
		AbonementObject: abonementObject,
	}

	return response, nil
}

func (c *AbonementgRPC) UpdateAbonement(g grpc.ClientStreamingServer[abonementProtobuf.UpdateAbonementRequest, abonementProtobuf.UpdateAbonementResponse]) error {
	abonementData, abonementPhoto, err := GetObjectData(
		&g,
		func(chunk *abonementProtobuf.UpdateAbonementRequest) interface{} {
			return chunk.GetAbonementDataForUpdate()
		},
		func(chunk *abonementProtobuf.UpdateAbonementRequest) []byte {
			return chunk.GetAbonementPhoto()
		},
	)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request data")
	}

	if abonementData == nil {
		logger.ErrorLogger.Printf("abonement data is empty")
		return status.Error(codes.InvalidArgument, "abonement data is empty")
	}

	castedAbonementData, ok := abonementData.(abonementProtobuf.AbonementDataForUpdate)
	if !ok {
		logger.ErrorLogger.Printf("abonement data is not of type AbonementProtobuf.AbonementDataForCreate")
		return status.Error(codes.InvalidArgument, "abonement data is not of type AbonementProtobuf.AbonementDataForCreate")
	}

	cmd := &dtos.UpdateAbonementCommand{
		Id:           uuid.MustParse(castedAbonementData.Id),
		Title:        castedAbonementData.Title,
		Validity:     castedAbonementData.Validity,
		VisitingTime: castedAbonementData.VisitingTime,
		Price:        int(castedAbonementData.Price),
		UpdatedTime:  time.Now(),
	}

	_, err = c.abonementUseCase.GetAbonementById(context.TODO(), uuid.MustParse(castedAbonementData.Id))
	if err != nil {
		return status.Error(codes.NotFound, "abonement not found")
	}

	var photoURL string
	var previousPhoto []byte
	if abonementPhoto != nil {
		previousPhoto, err = c.cloudUseCase.GetObjectByName(context.TODO(), "abonement/"+cmd.Id.String())
		if err != nil {
			logger.ErrorLogger.Printf("Failed to get previos photo from cloud: %v", err)
			return err
		}

		url, err := c.cloudUseCase.PutObject(context.TODO(), abonementPhoto, "abonement/"+cmd.Id.String())
		photoURL = url
		if err != nil {
			logger.ErrorLogger.Printf("Failed to create abonement photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create abonement photo in cloud")
		}
	}

	cmd.Photo = photoURL

	abonement, err := c.abonementUseCase.UpdateAbonement(context.TODO(), cmd)
	if err != nil {

		_, err := c.cloudUseCase.PutObject(context.TODO(), previousPhoto, "abonement/"+cmd.Id.String())
		if err != nil {
			logger.ErrorLogger.Printf("Failed to set previous photo in cloud: %v", err)
			return status.Error(codes.Internal, "Failed to create abonement photo in cloud")
		}

		return status.Error(codes.Internal, "Failed to create abonement")
	}

	abonementObject := &abonementProtobuf.AbonementObject{
		Id:           abonement.Id.String(),
		Title:        abonement.Title,
		Validity:     abonement.Validity,
		VisitingTime: abonement.VisitingTime,
		Photo:        abonement.Photo,
		Price:        int32(abonement.Price),
		CreatedTime:  abonement.CreatedTime.String(),
		UpdatedTime:  abonement.UpdatedTime.String(),
	}

	updateAbonementResponse := &abonementProtobuf.UpdateAbonementResponse{
		AbonementObject: abonementObject,
	}

	err = g.SendAndClose(updateAbonementResponse)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to send abonement update response: %v", err)
		return err
	}

	return nil
}

func (c *AbonementgRPC) DeleteAbonementById(ctx context.Context, request *abonementProtobuf.DeleteAbonementByIdRequest) (*abonementProtobuf.DeleteAbonementByIdResponse, error) {
	deletedAbonement, err := c.abonementUseCase.DeleteAbonementById(ctx, uuid.MustParse(request.Id))
	if err != nil {
		return nil, err
	}

	abonementObject := &abonementProtobuf.AbonementObject{
		Id:           deletedAbonement.Id.String(),
		Title:        deletedAbonement.Title,
		Validity:     deletedAbonement.Validity,
		VisitingTime: deletedAbonement.VisitingTime,
		Photo:        deletedAbonement.Photo,
		Price:        int32(deletedAbonement.Price),
		CreatedTime:  deletedAbonement.CreatedTime.String(),
		UpdatedTime:  deletedAbonement.UpdatedTime.String(),
	}

	deleteAbonementByIdResponse := &abonementProtobuf.DeleteAbonementByIdResponse{
		AbonementObject: abonementObject,
	}

	return deleteAbonementByIdResponse, nil
}

func (c *AbonementgRPC) GetAbonements(ctx context.Context, _ *emptypb.Empty) (*abonementProtobuf.GetAbonementsResponse, error) {

	abonementes, err := c.abonementUseCase.GetAbonementes(ctx)
	if err != nil {
		return nil, err
	}

	var abonementObjects []*abonementProtobuf.AbonementObject

	for _, abonement := range abonementes {

		abonementObject := &abonementProtobuf.AbonementObject{
			Id:           abonement.Id.String(),
			Title:        abonement.Title,
			Validity:     abonement.Validity,
			VisitingTime: abonement.VisitingTime,
			Photo:        abonement.Photo,
			Price:        int32(abonement.Price),
			CreatedTime:  abonement.CreatedTime.String(),
			UpdatedTime:  abonement.UpdatedTime.String(),
		}

		abonementObjects = append(abonementObjects, abonementObject)
	}

	response := &abonementProtobuf.GetAbonementsResponse{AbonementObjects: abonementObjects}

	return response, nil
}

func GetObjectData[T any, R any](
	g *grpc.ClientStreamingServer[T, R],
	extractObjectData func(chunk *T) interface{},
	extractObjectPhoto func(chunk *T) []byte,
) (interface{},
	[]byte,
	error,
) {
	var objectData interface{}
	var objectPhoto []byte

	for {
		chunk, err := (*g).Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.ErrorLogger.Printf("Error getting chunk: %v", err)
			return nil, nil, err
		}

		if ud := extractObjectData(chunk); ud != nil {
			objectData = ud
		}

		if uf := extractObjectPhoto(chunk); uf != nil {
			objectPhoto = append(objectPhoto, uf...)
		}
	}

	return objectData, objectPhoto, nil
}
