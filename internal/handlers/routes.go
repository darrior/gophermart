package handlers

import "github.com/go-chi/chi/v5"

func (s *server) setRoutes() {
	s.mux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.h.postAPIUserRegister)
		r.Post("/login", s.h.postAPIUserLogin)

		r.Group(func(r chi.Router) {
			r.Use(s.h.authMiddlware)

			r.Post("/orders", s.h.postAPIUserOrders)
			r.Post("balance/withdraw", s.h.postAPIUserBalanceWithdraw)
			r.Get("/orders", s.h.getAPIUserOrders)
			r.Get("/balance", s.h.getAPIUserBalance)
			r.Get("/withdrawals", s.h.getAPIUserWithdrawals)
		})
	})
}
