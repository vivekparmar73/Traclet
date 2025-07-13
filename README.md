# 📊 TrackIt – Personal Expense Tracker API

TrackIt is a beginner-friendly RESTful API built using **Golang**, **Gin**, and **GORM** that helps users track personal income and expenses.  
It supports basic user management (register/login), category-wise filtering, and monthly summaries.  
This project is designed for freshers or interns to learn backend API development using Go, MySQL, and Docker.

---

## 🛠️ Tech Stack

- **Golang** – Main backend language
- **Gin** – Web framework for handling routing and requests
- **GORM** – ORM to interact with MySQL database
- **MySQL** – Relational database for storing users and transactions
- **Docker** – Containerization and deployment
- **Postman** – API testing

---

## 📁 Project Structure
trackit/
├── config/ # Database configuration
├── controllers/ # Request handlers (user, expenses)
├── models/ # Database models
├── routes/ # All routes grouped here
├── main.go # Entry point
├── go.mod/go.sum # Go modules
├── Dockerfile # For containerization
├── .env # Environment variables


---

## 🚀 Features

- 📥 **User Registration/Login** (Basic DB check – no JWT)
- 💸 **Add Expenses** with category, amount, date, and notes
- 📂 **Category-wise Filtering**
- 📅 **Monthly Summaries**
- 🧹 Simple code structure for easy understanding
- 🐳 **Dockerized** for consistent deployment

---

