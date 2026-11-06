import { useState } from "react";
import "./MasterCabinet.css";

const monthNames = [
  "Январь", "Февраль", "Март", "Апрель", "Май", "Июнь",
  "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь",
];

const weekDays = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

function MasterCabinet() {
  const [tab, setTab] = useState("planner");
  const [month, setMonth] = useState(new Date(2026, 10, 1));
  const [selectedDate, setSelectedDate] = useState(4);

  const year = month.getFullYear();
  const monthIndex = month.getMonth();
  const daysInMonth = new Date(year, monthIndex + 1, 0).getDate();

  // Понедельник — первый день недели
  const firstDay = (new Date(year, monthIndex, 1).getDay() + 6) % 7;

  function changeMonth(offset) {
    setMonth(new Date(year, monthIndex + offset, 1));
  }

  function selectDay(day) {
    setSelectedDate(day);
    setTab("day");
  }

  return (
    <main className="master-app">
      <header className="master-header">
        <div className="master-brand">
          <span className="brand-mark">✓</span>
          <span>Slotvio</span>
        </div>
        {tab === "profile" && (
          <button className="settings-button" aria-label="Настройки профиля" title="Настройки профиля">
            ⚙
          </button>
        )}
      </header>

      {tab === "planner" && (
        <section className="master-content">
          <h1>Планер</h1>
          <p className="muted">Обзор записей и рабочего времени</p>

          <div className="calendar-card">
            <div className="calendar-heading">
              <button onClick={() => changeMonth(-1)} aria-label="Предыдущий месяц">
                ‹
              </button>
              <strong>{monthNames[monthIndex]} {year}</strong>
              <button onClick={() => changeMonth(1)} aria-label="Следующий месяц">
                ›
              </button>
            </div>

            <div className="calendar-grid">
              {weekDays.map((day) => (
                <span className="weekday" key={day}>{day}</span>
              ))}

              {Array.from({ length: firstDay }, (_, index) => (
                <span className="empty-day" key={`empty-${index}`} />
              ))}

              {Array.from({ length: daysInMonth }, (_, index) => {
                const day = index + 1;
                const hasAppointments = [4, 5, 8, 12, 14, 18, 21, 25].includes(day);

                return (
                  <button
                    key={day}
                    className={`calendar-day ${selectedDate === day ? "selected" : ""}`}
                    onClick={() => selectDay(day)}
                  >
                    <span>{day}</span>
                    {hasAppointments && <i />}
                  </button>
                );
              })}
            </div>
          </div>

          <section className="planner-settings">
            <h2>Доступность</h2>
            <button onClick={() => alert("Настройку рабочего времени добавим следующим шагом")}>
              Рабочие часы <span>›</span>
            </button>
            <button onClick={() => alert("Настройку срока записи добавим следующим шагом")}>
              Запись открыта на <span>30 дней ›</span>
            </button>
          </section>
        </section>
      )}

      {tab === "day" && (
        <section className="master-content">
          <h1>Мой день</h1>
          <p className="muted">{selectedDate} {monthNames[monthIndex].toLowerCase()} {year}</p>
          <div className="empty-state">
            <span>📅</span>
            <strong>Расписание на выбранный день</strong>
            <p>Здесь появятся записи клиентов.</p>
          </div>
          <button className="add-appointment" onClick={() => alert("Форму создания записи добавим следующим шагом")}>
            + Создать запись
          </button>
        </section>
      )}

      {tab === "profile" && (
        <section className="master-content">
          <h1>Профиль</h1>
          <div className="empty-state">
            <span>👤</span>
            <strong>Профиль мастера</strong>
            <p>Фото, имя, город, услуги и контакты добавим следующим шагом.</p>
          </div>
        </section>
      )}

      <nav className="bottom-nav">
        <button className={tab === "planner" ? "active" : ""} onClick={() => setTab("planner")}>
          <span>▦</span>Планер
        </button>
        <button className={tab === "day" ? "active" : ""} onClick={() => setTab("day")}>
          <span>▤</span>День
        </button>
        <button className={tab === "profile" ? "active" : ""} onClick={() => setTab("profile")}>
          <span>♙</span>Профиль
        </button>
      </nav>
    </main>
  );
}

export default MasterCabinet;
