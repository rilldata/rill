package server

// func (s *Server) user(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
// 	sess, _ := sessions.NewCookieStore([]byte(s.conf.SessionSecret)).Get(r, authSessionName)

// 	var profiles map[string]interface{}
// 	if sess.Values["profile"] == nil {
// 		http.Error(w, "Not Authenticated", http.StatusUnauthorized)
// 		return
// 	}
// 	profile := sess.Values["profile"].([]byte)
// 	err := json.Unmarshal(profile, &profiles)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = json.NewEncoder(w).Encode(profiles)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }
