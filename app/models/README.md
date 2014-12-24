## MODEL

   Model is to declare object for Database mapping.

   Because go using uppercase convention for public, exported fields whereas JSON prefers lower case syntax. As a result,
   it's strongly recommended to using tag such as "json:user_name" for json marshal/unmarshal to lower_case format.

##Samples:
    type User struct {
        Id           	int64	`json:"id"`
        Status			int8	`json:"status"`
        UserName 	 	string  `sql:"size:50" json:"user_name"`
        Email 		 	string  `sql:"size:50" json:"email"`
        Password 		string  `sql:"size:50"`
        HashedPassword  []byte	`json:"hashed_password"`
        FullName        string  `sql:"size:255" json:"full_name"`
        LastLogin    	time.Time `json:"last_login"`
        CreatedDate    	time.Time `json:"created_date"`
        UpdatedDate    	time.Time `json:"updated_date"`
        DeletedDate    	time.Time `json:"deleted_date"`
    }

