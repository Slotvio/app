import { useState } from "react";
import "./ClientCabinet.css";

const masters = [
  {
    id: 1,
    name: "Анна Смирнова",
    direction: "Маникюр",
    city: "Минск",
    format: "В студии",
    price: 60,
    next: "Сегодня, 16:30",
    initials: "АС",
  },
  {
    id: 2,
    name: "Мария Ковалёва",
    direction: "Брови",
    city: "Минск",
    format: "На дому",
    price: 45,
    next: "Завтра, 11:00",
    initials: "МК",
  },
  {
    id: 3,
    name: "Екатерина Ли",
    direction: "Маникюр",
    city: "Гродно",
    format: "С выездом",
    price: 75,
    next: "Пятница, 14:00",
    initials: "ЕЛ",
  },
];

function ClientCabinet() {
  const [tab, setTab] = useState("search");
  const [query, setQuery] = useState("");
  const [direction, setDirection] = useState("");
  const [city, setCity] = useState("");
  const [format, setFormat] = useState("");
  const [maxPrice, setMaxPrice] = useState("");

  const filteredMasters = masters.filter((master) => {
    const matchesQuery =
      `${master.name} ${master.direction}`.toLowerCase().includes(query.toLowerCase());

    return (
      matchesQuery &&
      (!direction || master.direction === direction) &&
      (!city || master.city === city) &&
      (!format || master.format === format) &&
      (!maxPrice || master.price <= Number(maxPrice))
    );
  });

  return (
    <main className="client-app">
      <header className="client-header">
        <div className="client-brand">
          <span className="client-brand-mark">✓</span>
          <span>Slotvio</span>
        </div>
        <button className="client-avatar" onClick={() => setTab("profile")}>
          А
        </button>
      </header>

      {tab === "search" && (
        <section>
          <h1>Найдите своего мастера</h1>
          <p className="client-muted">Выберите услугу и удобное время</p>

          <input
            className="master-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Мастер или услуга"
          />

          <div className="client-filters">
            <select value={direction} onChange={(event) => setDirection(event.target.value)}>
              <option value="">Направление</option>
              <option>Маникюр</option>
              <option>Брови</option>
            </select>

            <select value={city} onChange={(event) => setCity(event.target.value)}>
              <option value="">Город</option>
              <option>Минск</option>
              <option>Гродно</option>
            </select>

            <select value={format} onChange={(event) => setFormat(event.target.value)}>
              <option value="">Формат работы</option>
              <option>В студии</option>
              <option>На дому</option>
              <option>С выездом</option>
            </select>

            <input
              type="number"
              min="0"
              value={maxPrice}
              onChange={(event) => setMaxPrice(event.target.value)}
              placeholder="Цена до, BYN"
            />
          </div>

          <div className="client-results-heading">
            <h2>Мастера рядом</h2>
            <span>{filteredMasters.length}</span>
          </div>

          <div className="master-list">
            {filteredMasters.map((master) => (
              <article className="master-card" key={master.id}>
                <div className="master-photo">{master.initials}</div>
                <div className="master-info">
                  <h3>{master.name}</h3>
                  <p>{master.direction} · {master.city}</p>
                  <p>{master.format} · от {master.price} BYN</p>
                  <span className="master-next">Ближайшая запись: {master.next}</span>
                </div>
                <button
                  className="master-open"
                  onClick={() => alert(`Профиль мастера: ${master.name}`)}
                  aria-label={`Открыть профиль ${master.name}`}
                >
                  ›
                </button>
              </article>
            ))}
          </div>
        </section>
      )}

      {tab === "bookings" && (
        <section>
          <h1>Мои записи</h1>
          <p className="client-muted">Предстоящие, история и отменённые записи</p>
          <div className="client-placeholder">Здесь будут ваши записи</div>
        </section>
      )}

      {tab === "profile" && (
        <section>
          <h1>Профиль</h1>
          <div className="client-profile-card">
            <div className="client-profile-photo">А</div>
            <h2>Александр</h2>
            <p>@telegram</p>
            <button onClick={() => alert("Редактирование профиля добавим позже")}>
              Редактировать профиль
            </button>
          </div>
        </section>
      )}

      <nav className="client-bottom-nav">
        <button className={tab === "search" ? "active" : ""} onClick={() => setTab("search")}>
          <span>⌕</span>Поиск
        </button>
        <button className={tab === "bookings" ? "active" : ""} onClick={() => setTab("bookings")}>
          <span>▤</span>Мои записи
        </button>
        <button className={tab === "profile" ? "active" : ""} onClick={() => setTab("profile")}>
          <span>♙</span>Профиль
        </button>
      </nav>
    </main>
  );
}

export default ClientCabinet;