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

	e.GET("/getallpost", authController.GetAllPost)                       //list all post (include post content and user info)
	e.GET("/searchpost/:name", authController.SeachPost)                  //searchpost by name
	e.GET("/getallpost/:username", authController.GetAllPostByUsername)   //get all post by username
	e.GET("/getpostbypostid/:id", authController.GetPostByPostID)         //หน้าดู post คนอื่น
	e.GET("/userprofile/:id", authController.GetUserProfileByID)          //หน้าดู profile คนอื่นจาก user profile
	e.GET("/getcommentsbypostid/:id", authController.GetCommentsByPostID) //หน้าดู comment ของ post นั้นๆ
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
	api.POST("/createpost", authController.CreatePost)               //หน้า create post
	api.GET("/mypost", authController.GetAllMyPost)                  //หน้า my profile
	api.PUT("/editpost/:id", authController.EditPostByPostID)        ////หน้า edit post (เข้าได้ผ่าน post_id)
	api.DELETE("/deletepost/:id", authController.DeletePostbyPostID) //หน้าลบ post ของตัวเอง(เข้าได้ผ่าน post_id)

	// api.POST("/user/ratepost", authController.RatePost)

	//api.GET("/mypost", authController.GetAllMyPost)         //หน้าดู post ทั้งหมดของฉัน          //get post ทั้งหมดของฉัน (เช็ค user_id จาก token)
	//api.GET("/getallpost", authController.GetAllPost)       //หน้าดูpost ทั้งหมดของทุกคน

	//favorite
	api.POST("/favoritepost/:postId", authController.FavoritePost)
	api.DELETE("/unfavoritepost/:postId", authController.UnFavoritePost)
	api.GET("/getallfavoritepost", authController.GetAllFavoritePost)

	//ratings
	api.POST("/ratepost", authController.RatePost)                //ratepost
	api.GET("/getratepost", authController.GetRatePost)           //ไม่ได้ใช้ เอามาเทสดู post ทั้วหมดที่โดน rate เฉยๆ
	api.GET("/getratescore/:postId", authController.GetRateScore) //ใช้บอกว่ารอบที่แล้ว user ให้คะแนนโพสกี่คะแนน

	//comment
	api.GET("/getcommentsbypostid/:id", authController.GetCommentsByPostID)
	api.POST("/addcomment", authController.AddComment)
	api.PUT("/editcomment", authController.EditComment)
	api.DELETE("/deletecommentbycommentid/:id", authController.DeleteCommentByCommentID)

	//follow
	api.POST("/followuser", authController.FollowUser)                          //ใช้ตอนกดฟอลคนอื่น
	api.DELETE("/unfollowuser/:id", authController.UnFollowUser)                //ใช้ตอนกดอันฟอลคนอื่น
	e.GET("/getallfollower/:id", authController.GetAllFollower)                 //ใช้นับจำนวนคนฟอลทั้งหมด
	e.GET("/getallfollowing/:id", authController.GetAllFollowing)               //ใช้นับจำนวนคนที่ฟอลทั้งหมด
	api.GET("/getallfollowerdetail/:id", authController.GetAllFollowerDetail)   //ใช้บอกข้อมูลคนฟอล รูปโปรโฟล์ ชื่อ　นามสกุล
	api.GET("/getallfollowingdetail/:id", authController.GetAllFollowingDetail) //ใช้บอกข้อมูลคนที่ฟอล รูปโปรโฟล์ ชื่อ　นามสกุล

	//badges
	e.GET("/getallbadges/:id", authController.GetAllBadges) //ใช้บอกbadges ทั้งหมดที่ id นั้นมี

}
