package main

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"time"
	desc "valera/pkg/chat_v1"
)

const grpcPort = 50051

type server struct {
	desc.UnimplementedChatAPIServer
}

var db *sql.DB

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	connStr := fmt.Sprintf(
		"host=localhost "+
			"port=%s "+
			"user=%s "+
			"password=%s "+
			"dbname=%s "+
			"sslmode=disable",
		os.Getenv("PG_PORT"),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_DATABASE_NAME"),
	)

	var dbErr error
	db, dbErr = sql.Open("postgres", connStr)
	if dbErr != nil {
		log.Fatalf("failed to open database connection: %v", dbErr)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	fmt.Println("Database connection established successfully")
}

func (s *server) Create(ctx context.Context, request *desc.CreateRequest) (*desc.CreateResponse, error) {
	for _, username := range request.Usernames {
		builderInsert := sq.Insert("usernames").
			PlaceholderFormat(sq.Dollar).
			Columns("usernames").
			Values(username).
			Suffix("RETURNING id")

		query, args, err := builderInsert.ToSql()
		if err != nil {
			log.Println("failed to build query: %v", err)
		}

		var noteID int
		err = db.QueryRow(query, args...).Scan(&noteID)
		if err != nil {
			log.Println("failed to insert username: %v", err)
		}

		log.Printf("inserted username with id: %d", noteID)
	}
	fmt.Println("aloha!")
	return &desc.CreateResponse{
		Id: gofakeit.Int64(),
	}, nil
}

func (s *server) Delete(ctx context.Context, request *desc.DeleteRequest) (*desc.DeleteResponse, error) {
	if request.Id == 0 {
		return nil, fmt.Errorf("invalid ID provided")
	}

	builderInsert := sq.Delete("usernames").PlaceholderFormat(sq.Dollar).Where(sq.Eq{"id": request.Id})
	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	log.Printf("Executing query: %s with args: %v", query, args)

	_, err = db.Exec(query, args...)
	if err != nil {
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to delete record: %w", err)
	}

	return &desc.DeleteResponse{}, nil
}

func (s *server) SendMessage(ctx context.Context, request *desc.SendMessageRequest) (*desc.SendMessageResponse, error) {
	fmt.Println(request.Text, request.Timestamp, request.From)
	builderInsert := sq.Insert("chat").
		PlaceholderFormat(sq.Dollar).
		Columns("fromm", "msg", "time").
		Values(request.From, request.Text, time.Now())
	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to build query: %v", err)
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to delete record: %w", err)
	}

	log.Printf("Executing query: %s with args: %v", query, args)
	return &desc.SendMessageResponse{
		Empty: &emptypb.Empty{},
	}, nil
}

func main() {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatAPIServer(s, &server{})

	log.Printf("server listening at %v", list.Addr())

	if err = s.Serve(list); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
