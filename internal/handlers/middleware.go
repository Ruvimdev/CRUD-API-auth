package handlers

import (
	"context"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string 
const UserIDKey contextKey = "userID"

   //сюда передали ключ			      //эта функция создает и возвращает мидлвару
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	//сюда просто прокидываем next
	return func(next http.Handler) http.Handler { 
		//сама работа с http, http.handlerfunc(func() {})
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Достаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, http.StatusUnauthorized, "отсуствует токен авторизации")
				return 
			}

			 //bearer тип токена
			 // Обычно токен передают в формате "Bearer <сам_токен>"
             // Отрываем слово Bearer
			 // Берем строку: "Bearer eyJhbGci..." и рубим её по пробелу
			headerParts := strings.Split(authHeader, " ")//массив [0] bearer, [1] token
			//масив из 2х частей и первое слово bearer?
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				sendError(w, http.StatusUnauthorized, "неверный формат токена")
				return 
			}	
			tokenString := headerParts[1] //забрали массив с токеном

			// 2. парсим и проверяем токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
										
				// "переменная.(ОжидаемыйТип)" утверждение типа				
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					//возвращает два значения, саму переменную и булевое значение (true совпал)	
				//если ok == false (алгоритм не совпал), это хак. атака
					return nil, http.ErrAbortHandler
				}

				return []byte(jwtSecret), nil 
			})

			if err != nil || !token.Valid {
				sendError(w, http.StatusUnauthorized, "недействительный или просроченный токен")
				return 
			}

			//3. достаем user_id из токена | claims.() - мапа 
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				sendError(w, http.StatusUnauthorized, "ошибка чтения токена")
				return 
			}
			//float64, json числа по умолчанию парсятся как float
			//обращаемся к мапе по ключу: claims["user_id"]
			rawID, ok := claims["user_id"]
			if !ok {
				sendError(w, http.StatusUnauthorized, "невалидный токен")
				return 
			}

			//4. кладем id юзера в контекст
			// берем старый контекст, кладем туда userID и получаем новый контекст
			ctx := context.WithValue(r.Context(), UserIDKey, rawID)

			//5. пропускаем запрос дальше, отдаем новый контекст с айди
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}  