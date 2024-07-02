package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/bloomingFlower/rssagg/internal/database"
	api "github.com/bloomingFlower/rssagg/protos"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type server struct {
	api.UnimplementedApiServiceServer
	DB *database.Queries
}

func (s *server) Healthz(ctx context.Context, req *api.HealthzRequest) (*api.HealthzResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthz not implemented")
}

//type apiConfig struct {
//	DB *database.Queries
//}

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 요청을 받을 때마다 로그를 출력합니다.
	log.Printf("Received gRPC request - Method: %s", info.FullMethod)
	// 다음 핸들러를 호출합니다.
	return handler(ctx, req)
}

func loggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// 스트림 요청을 받을 때마다 로그를 출력합니다.
	log.Printf("Received gRPC stream request - Method: %s", info.FullMethod)
	// 다음 핸들러를 호출합니다.
	return handler(srv, ss)
}

func main() {
	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Can't connect to database:", err)
	}

	db := database.New(conn)
	//apiCfg := apiConfig{
	//	DB: db,
	//}

	go startScraping(db, 60, time.Minute)
	//
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := &server{
		UnimplementedApiServiceServer: api.UnimplementedApiServiceServer{},
		DB:                            db,
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				loggingInterceptor, // 추가할 Unary 인터셉터
				s.middlewareAuth,   // 기존 Unary 인터셉터
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				s.middlewareAuthStream,   // 기존 Stream 인터셉터
				loggingStreamInterceptor, // 추가할 Stream 인터셉터
			),
		),
	)
	reflection.Register(grpcServer)
	api.RegisterApiServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	//
	//router := chi.NewRouter()
	//
	//router.Use(cors.Handler(cors.Options{
	//	AllowedOrigins:   []string{"https://*", "http://*"},
	//	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	//	AllowedHeaders:   []string{"*"},
	//	ExposedHeaders:   []string{"Link"},
	//	AllowCredentials: false,
	//	MaxAge:           300,
	//}))
	//
	//v1Router := chi.NewRouter()
	//v1Router.Get("/healthz", handlerReadiness)
	//v1Router.Get("/err", handlerErr)
	//
	//v1Router.Post("/users", apiCfg.handlerCreateUser)
	//v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.handlerGetUser))
	//
	//v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	//v1Router.Get("/feeds", apiCfg.handlerGetFeeds)
	//
	//v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerCreateFeedFollow))
	//v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.handlerGetFeedFollows))
	//v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeedFollows))
	//
	//v1Router.Get("/posts", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))
	//
	////v1Router.Get("/view", apiCfg.middlewareAuth(apiCfg.handlerView))
	//v1Router.Get("/view", apiCfg.handlerView)
	//
	//router.Mount("/v1", v1Router)
	//
	//srv := &http.Server{
	//	Handler: router,
	//	Addr:    ":" + portString,
	//}
	//
	//log.Printf("Server starting on port %v", portString)
	//log.Fatal(srv.ListenAndServe())
}
