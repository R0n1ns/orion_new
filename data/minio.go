package data

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Manager инкапсулирует клиента MinIO и имя бакета.
type Manager struct {
	Client *minio.Client
	Bucket string
}

var MinioMgr *Manager

//var (
//	endpoint  string = "localhost:9000"
//	accessKey string = "youraccesskey"
//	secretKey string = "yoursecretkey"
//	useSSL    bool   = false
//	bucket    string = "images"
//)

func init() {
	var err error

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	bucket := os.Getenv("MINIO_BUCKET")

	// инициализация может происходить так (значения берутся из конфига или переменных окружения):
	//MinioMgr, err = NewManager("minio:9000", "youraccesskey", "yoursecretkey", "images", false)
	// Инициализация клиента.
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: useSSL})
	if err != nil {
		fmt.Errorf("failed to initialize minio client: %w", err)
	}

	// Проверка существования бакета.
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		fmt.Errorf("error checking bucket existence: %w", err)
	}
	if !exists {
		// Создание бакета.
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	MinioMgr = &Manager{
		Client: client,
		Bucket: bucket,
	}
	//if err != nil {
	//	fmt.Errorf("Ошибка подключения к minio")
	//}
	fmt.Println("Минио подключен успешно")
}

// UploadImage загружает данные изображения в указанный бакет.
func UploadImage(ctx context.Context, objectName string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := MinioMgr.Client.PutObject(ctx, MinioMgr.Bucket, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}
	return nil
}

// GetPhoto возвращает фотографию в виде data URL, полученную из MinIO.
// Файл ищется по пути "images/{hash}.jpg". Если файл не найден, возвращается строка "none".
//
// Параметры:
//   - hash: строка-хэш, которая используется для формирования имени файла.
//
// Возвращаемое значение:
//   - string: data URL изображения или "none", если изображение не найдено.
func GetPhoto(hash string) string {
	objectName := hash + ".jpg"
	ctx := context.Background()

	// Получаем объект из MinIO.
	obj, err := MinioMgr.Client.GetObject(ctx, MinioMgr.Bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return "none"
	}

	// Читаем содержимое объекта.
	imageBytes, err := io.ReadAll(obj)
	if err != nil || len(imageBytes) == 0 {
		return "none"
	}

	// Кодируем изображение в base64 и формируем data URL.
	encodedImage := base64.StdEncoding.EncodeToString(imageBytes)
	profilePictureURL := "data:image/jpeg;base64," + encodedImage
	return profilePictureURL
}
