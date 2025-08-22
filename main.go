package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func main() {
	app := fiber.New()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, Railway!",
		})
	})

	// Add new upload endpoint
    app.Post("/upload", func(c *fiber.Ctx) error {
        return handlePost(c)
    })

	app.Listen(getPort())
}

func uploadToCloudinary(filePath string, filename string) (string, string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	const namespace = "editor/"
	publicID := namespace + filename

	cld, _ := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)

	// Upload the my_picture.jpg image and set the PublicID to "my_image". 
	var ctx = context.Background()
	_, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{PublicID: publicID});
	if err != nil {
		fmt.Println("error uploading image:", err)
		return "", "", err
	}

	// Get details about the image with PublicID "my_image" and log the secure URL.
	resp, err := cld.Admin.Asset(ctx, admin.AssetParams{PublicID: publicID});
	if err != nil {
		fmt.Println("error")
	}
	log.Println(resp.SecureURL)

	// Instantiate an object for the asset with public ID "my_image"
	myImage, err := cld.Image(publicID)
	if err != nil {
		fmt.Println("error")
	}

	// Add the transformation
	myImage.Transformation = "c_fill,h_500,w_500"

	// Generate and print the delivery URL
	url, err := myImage.String()

	log.Println(url)
	return url, publicID, err
}

func handlePost(c *fiber.Ctx) error {
	// Get file from request
	file, err := c.FormFile("image"); if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No image provided",
		})
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "./uploads"
	if errMkdir := os.MkdirAll(uploadsDir, 0755); errMkdir != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create uploads directory",
		})
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", c.Context().ID(), "img")
	filepath := filepath.Join(uploadsDir, filename)

	// Save file
	if errSaveFile := c.SaveFile(file, filepath); errSaveFile != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save image",
		})
	}

	// Schedule file deletion for when this function completes
	defer func() {
		if err := os.Remove(filepath); err != nil {
			log.Printf("Failed to delete local file %s: %v", filepath, err)
		} else {
			log.Printf("Successfully deleted local file: %s", filepath)
		}
	}()

	url, publicID, errUpload := uploadToCloudinary(filepath, filename); if errUpload != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload image to Cloudinary",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"filename": url,
		"public_id": publicID,
	})
}