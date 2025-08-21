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
		log.Fatal("Error loading .env file")
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, Railway!",
		})
	})

	// Add new upload endpoint
    app.Post("/upload", func(c *fiber.Ctx) error {
        // Get file from request
        file, err := c.FormFile("image"); if err != nil {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "No image provided",
            })
        }

        // Create uploads directory if it doesn't exist
        uploadsDir := "./uploads"
        if err := os.MkdirAll(uploadsDir, 0755); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to create uploads directory",
            })
        }

        // Generate unique filename
        filename := fmt.Sprintf("%d_%s", c.Context().ID(), filepath.Base(file.Filename))
        filepath := filepath.Join(uploadsDir, filename)

        // Save file
        if err := c.SaveFile(file, filepath); err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "Failed to save image",
            })
        }

		url, err := uploadToCloudinary(filepath); if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload image to Cloudinary",
			})
		}

        // Return success response
        return c.Status(fiber.StatusOK).JSON(fiber.Map{
            "success": true,
            "filename": url,
        })
    })

	app.Listen(getPort())
}

func uploadToCloudinary(filePath string) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	cld, _ := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)

	// Upload the my_picture.jpg image and set the PublicID to "my_image". 
	var ctx = context.Background()
	_, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{PublicID: "my_image"});
	if err != nil {
		fmt.Println("error uploading image:", err)
		return "", err
	}

	// Get details about the image with PublicID "my_image" and log the secure URL.
	resp, err := cld.Admin.Asset(ctx, admin.AssetParams{PublicID: "my_image"});
	if err != nil {
		fmt.Println("error")
	}
	log.Println(resp.SecureURL)

	// Instantiate an object for the asset with public ID "my_image"
	myImage, err := cld.Image("my_image")
	if err != nil {
		fmt.Println("error")
	}

	// Add the transformation
	myImage.Transformation = "c_fill,h_250,w_250"

	// Generate and print the delivery URL
	url, err := myImage.String()

	log.Println(url)
	return url, err
}
