package main

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"os"
	"time"
	desc "valera/pkg/user_v1"
)

const grpcPort = 50052

type server struct {
	desc.UnimplementedUserAPIServer
}

var db *sql.DB

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error parse .env file", err)
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

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	log.Println("Note id:", req.GetId())
	buildGet := sq.Select("id", "name", "email", "password", "role", "created_at").
		From("auth").
		Where(sq.Eq{"id": req.Id}).
		PlaceholderFormat(sq.Dollar)
	query, args, err := buildGet.ToSql()
	if err != nil {
		log.Println("failed to build query: %v", err)
	}

	var id int64
	var name, email, password, role string
	var created_at time.Time
	err = db.QueryRow(query, args...).Scan(&id, &name, &email, &password, &role, &created_at)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("no user found with ID: %d", req.GetId())
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	roleValueMap := map[string]desc.User_Role{
		"0": desc.User_user,
		"1": desc.User_admin,
	}

	roleValue, ok := roleValueMap[role]
	if !ok {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	return &desc.GetResponse{
		User: &desc.User{
			Id:        id,
			Name:      name,
			Email:     email,
			Role:      roleValue,
			CreatedAt: timestamppb.New(created_at),
		},
	}, nil
}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	builderInsert := sq.Insert("auth").
		Columns("name", "email", "password", "role").
		Values(req.Info.Name, req.Info.Email, req.Info.Password, req.Info.Role).
		PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING id")
	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	var id int64
	err = db.QueryRow(query, args...).Scan(&id)
	return &desc.CreateResponse{Id: id}, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	builderUpdate := sq.Update("auth").Where(sq.Eq{"id": req.Id}).PlaceholderFormat(sq.Dollar)
	if name := req.GetName(); name != nil {
		builderUpdate = builderUpdate.Set("name", name.GetValue())
	}
	if email := req.GetEmail(); email != nil {
		builderUpdate = builderUpdate.Set("email", email.GetValue())
	}
	query, args, err := builderUpdate.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	_, err = db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	builderDelete := sq.Delete("auth").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id})
	query, args, err := builderDelete.ToSql()
	if err != nil {
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}
	_, err = db.Exec(query, args...)
	if err != nil {
		log.Printf("failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to delete record: %w", err)
	}
	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserAPIServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
