package routes

import (
	"go-auth/controllers"
	jwtmiddleware "go-auth/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Init(e *echo.Echo, authController *controllers.AuthController) {
	// CORS middleware for frontend integration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, // In production, specify your frontend domain
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Traditional auth routes
	e.POST("/register", authController.Register)
	e.POST("/login", authController.Login)

	// Google OAuth routes
	e.GET("/auth/google", controllers.GoogleLogin)
	e.GET("/auth/google/callback", controllers.GoogleCallback)

	// Protected routes
	api := e.Group("/api")
	api.Use(jwtmiddleware.JWTMiddleware())

	// api.GET("/profile", func(c echo.Context) error {
	// 	userID := c.Get("user_id")
	// 	return c.JSON(200, echo.Map{
	// 		"message": "You are logged in",
	// 		"user_id": userID,
	// 	})
	// })

	// Get current user info
	// api.GET("/me", controllers.GetCurrentUser)
	//api.GET("/userprofile", controllers.GetUserProfile)
	api.GET("/getalluser", authController.GetAllUser)                //หน้า show user profile (ของเรา)
	api.GET("/userprofile", authController.GetUserProfile)           //หน้า show user profile (ของเรา)
	api.GET("/userprofile/:id", authController.GetUserProfileByID)   //หน้า show user profile (ของเรา)
	api.PUT("/userprofile", authController.UpdateUserProfile)        //หน้า edit user profile
	api.POST("/createpost", authController.CreatePost)               //หน้า post
	api.PUT("/editpost/:id", authController.EditPostByPostID)        ////หน้า edit post (เข้าได้ผ่าน post_id)
	api.DELETE("/deletepost/:id", authController.DeletePostbyPostID) //หน้าลบ post ของตัวเอง(เข้าได้ผ่าน post_id)

	// api.POST("/user/ratepost", authController.RatePost)
	api.GET("/getpost/:id", authController.GetPostByPostID) //หน้าดู post คนอื่น            //get post จาก post id
	api.GET("/mypost", authController.GetAllMyPost)         //หน้าดู post ทั้งหมดของฉัน          //get post ทั้งหมดของฉัน (เช็ค user_id จาก token)
	api.GET("/getallpost", authController.GetAllPost)       //หน้าดูpost ทั้งหมดของทุกคน
}
