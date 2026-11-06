import { useState } from "react";
import "./App.css";
import MasterCabinet from "./MasterCabinet";
import ClientCabinet from "./ClientCabinet";

function App() {
  //return <ClientCabinet />;
  return <MasterCabinet />;
  const [role, setRole] = useState(null);
  const [mode, setMode] = useState("register");

  function chooseRole(selectedRole) {
    setRole(selectedRole);
  }

  function goBack() {
    setRole(null);
  }

  return (
    <main className="page">
      <section className="auth-card">
        <header className="brand">
          <div className="brand-icon">✓</div>
          <span>Slotvio</span>
        </header>

        {!role ? (
          <>
            <h1>Добро пожаловать!</h1>
            <p className="subtitle">Сервис онлайн-записи</p>

            <h2 className="section-title">Продолжить как</h2>

            <div className="role-list">
              <button
                className="role-card"
                onClick={() => chooseRole("master")}
              >
                <span className="role-icon">✂</span>
                <span className="role-name">Мастер</span>
                <span className="role-description">
                  Управляйте услугами и записями
                </span>
              </button>

              <button
                className="role-card"
                onClick={() => chooseRole("client")}
              >
                <span className="role-icon">♙</span>
                <span className="role-name">Клиент</span>
                <span className="role-description">
                  Находите мастеров и записывайтесь
                </span>
              </button>
            </div>
          </>
        ) : (
          <>

            <h1>{role === "master" ? "Кабинет мастера" : "Добро пожаловать!"}</h1>
            <p className="subtitle">
              {mode === "register" ? "Создайте аккаунт" : "Войдите в аккаунт"}
            </p>

            <div className="mode-switch">
              <button
                className={mode === "register" ? "active" : ""}
                onClick={() => setMode("register")}
              >
                Регистрация
              </button>
              <button
                className={mode === "login" ? "active" : ""}
                onClick={() => setMode("login")}
              >
                Вход
              </button>
            </div>

            <form className="auth-form" onSubmit={(event) => event.preventDefault()}>
              {mode === "register" && (
                <label>
                  Ваше имя
                  <input type="text" placeholder="Например, Иван" />
                </label>
              )}

              <label>
                Электронная почта
                <input type="email" placeholder="name@example.com" />
              </label>

              <label>
                Пароль
                <input type="password" placeholder="Не менее 8 символов" />
              </label>

              <button className="primary-button" type="submit">
                {mode === "register" ? "Создать аккаунт" : "Войти"}
              </button>
              
              <button className="back-button" onClick={goBack}>
              ← Назад
              </button>
            </form>
          </>
        )}
      </section>
    </main>
  );
}

export default App;