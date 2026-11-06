
import { useState } from "react";
import "./ClientCabinet.css";

const masters = [
  {
    id: 1,
    name: "Евгения",
    services: ["Маникюр", "Педикюр"],
    city: "Фрунзенский, Минск",
    price: 60,
    rating: 5.0,
    image: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=700",
  },
  {
    id: 2,
    name: "Анна Смирнова",
    services: ["Массаж", "Уход за телом"],
    city: "Центральный, Минск",
    price: 80,
    rating: 4.9,
    image: "https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=700",
  },
  {
    id: 3,
    name: "Мария",
    services: ["Брови", "Ламинирование"],
    city: "Гродно",
    price: 45,
    rating: 5.0,
    image: "https://images.unsplash.com/photo-1531123897727-8f129e1688ce?w=700",
  },
];

const cities = {
  Беларусь: {
    popular: ["Минск", "Гомель", "Гродно", "Витебск", "Могилёв", "Брест", "Бобруйск", "Барановичи"],
    nearby: {
      Минск: ["Дзержинск", "Фаниполь", "Заславль", "Смолевичи"],
      Гомель: ["Добруш", "Ветка"],
      Гродно: ["Скидель", "Сопоцкин"],
    },
  },
  Россия: {
    popular: ["Москва", "Санкт-Петербург", "Казань", "Сочи", "Екатеринбург", "Новосибирск"],
    nearby: {},
  },
};

function ClientCabinet() {
  const [tab, setTab] = useState("search");
  const [query, setQuery] = useState("");
  const [country, setCountry] = useState("Беларусь");
  const [city, setCity] = useState("");
  const [citySearch, setCitySearch] = useState("");
  const [cityOpen, setCityOpen] = useState(false);
  const [dateOpen, setDateOpen] = useState(false);
  const [selectedDates, setSelectedDates] = useState([]);
  const [datePreset, setDatePreset] = useState("any");
  const [timeOfDay, setTimeOfDay] = useState("any");
  const [exactTime, setExactTime] = useState("");
  const [appliedDateFilter, setAppliedDateFilter] = useState({
    selectedDates: [],
    datePreset: "any",
    timeOfDay: "any",
    exactTime: "",
  });
  const [calendarMonth, setCalendarMonth] = useState(() => {
    const today = new Date();
    return new Date(today.getFullYear(), today.getMonth(), 1);
  });
  const [selectedMaster, setSelectedMaster] = useState(null);
  const [bookingTab, setBookingTab] = useState("upcoming");

  const today = new Date();
  const todayKey = toDateKey(today);
  const calendarYear = calendarMonth.getFullYear();
  const calendarMonthIndex = calendarMonth.getMonth();
  const firstWeekday = (new Date(calendarYear, calendarMonthIndex, 1).getDay() + 6) % 7;
  const daysInMonth = new Date(calendarYear, calendarMonthIndex + 1, 0).getDate();
  const monthKey = `${calendarYear}-${String(calendarMonthIndex + 1).padStart(2, "0")}`;
  const currentMonthKey = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}`;

  function changeMonth(offset) {
    setCalendarMonth((current) => new Date(current.getFullYear(), current.getMonth() + offset, 1));
  }

  function choosePreset(preset) {
    setDatePreset(preset);
    if (preset === "any") {
      setSelectedDates([]);
      return;
    }

    const base = new Date(today.getFullYear(), today.getMonth(), today.getDate());
    if (preset === "today") {
      setPresetDates([base]);
    } else if (preset === "tomorrow") {
      base.setDate(base.getDate() + 1);
      setPresetDates([base]);
    } else {
      const daysUntilSaturday = (6 - base.getDay() + 7) % 7;
      const saturday = new Date(base);
      saturday.setDate(base.getDate() + daysUntilSaturday);
      const sunday = new Date(saturday);
      sunday.setDate(saturday.getDate() + 1);
      setPresetDates([saturday, sunday]);
    }
  }

  function setPresetDates(dates) {
    setSelectedDates(dates.map(toDateKey));
    setCalendarMonth(new Date(dates[0].getFullYear(), dates[0].getMonth(), 1));
  }

  function toggleDate(dateKey) {
    const nextDates = selectedDates.includes(dateKey)
      ? selectedDates.filter((item) => item !== dateKey)
      : [...selectedDates, dateKey].sort();
    setDatePreset(nextDates.length ? null : "any");
    setSelectedDates(nextDates);
  }

  function resetDateFilter() {
    setDatePreset("any");
    setSelectedDates([]);
    setTimeOfDay("any");
    setExactTime("");
  }

  const dateFilterLabel = formatDateFilterLabel(appliedDateFilter);

  const filteredMasters = masters.filter((master) => {
    const text = `${master.name} ${master.services.join(" ")} ${master.city}`.toLowerCase();
    const matchesQuery = text.includes(query.toLowerCase());
    const matchesCity = !city || master.city.toLowerCase().includes(city.toLowerCase());
    return matchesQuery && matchesCity;
  });

  const nearbyCities = cities[country].nearby[city] || [];

  return (
    <main className="client-app">
      <header className="client-header">
        <div className="client-brand">
          <span className="client-brand-mark">✓</span>
          <span>Slotvio</span>
        </div>
      </header>

      {tab === "search" && !selectedMaster && (
        <section>
          <h1 className="client-title">Найдите своего мастера</h1>
          <p className="client-subtitle">
            Выбирайте специалистов и записывайтесь онлайн
          </p>

          <div className="search-panel">
            <input
              className="master-search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Опишите, что ищете..."
            />

            <div className="search-controls">
              <div className="city-picker">
                <button
                  className="picker-button"
                  aria-expanded={cityOpen}
                  onClick={() => {
                    setDateOpen(false);
                    setCityOpen((open) => !open);
                  }}
                >
                  <span>⌖</span>
                  {city || country}
                  <span className="picker-arrow">⌄</span>
                </button>

                <button
                  className="picker-button date-button"
                  aria-expanded={dateOpen}
                  onClick={() => {
                    setCityOpen(false);
                    setDateOpen((open) => !open);
                  }}
                >
                  <span>▦</span>
                  <span className="date-button-label">{dateFilterLabel}</span>
                  <span className="picker-arrow">⌄</span>
                </button>

                <button className="find-button" onClick={() => {
                  setCityOpen(false);
                  setDateOpen(false);
                }}>
                  Найти
                </button>
              </div>

              {dateOpen && (
                <section className="date-dropdown" aria-label="Фильтр даты и времени">
                  <div className="date-popover-heading">
                    <h2>Дата и время</h2>
                    <button type="button" className="popover-close" aria-label="Закрыть календарь" onClick={() => setDateOpen(false)}>×</button>
                  </div>

                  <div className="date-presets" aria-label="Быстрый выбор даты">
                    {[
                      ["any", "Любой день"],
                      ["today", "Сегодня"],
                      ["tomorrow", "Завтра"],
                      ["weekends", "Выходные"],
                    ].map(([value, label]) => (
                      <button
                        type="button"
                        key={value}
                        className={datePreset === value ? "selected" : ""}
                        aria-pressed={datePreset === value}
                        onClick={() => choosePreset(value)}
                      >
                        {label}
                      </button>
                    ))}
                  </div>

                  <div className="date-calendar-heading">
                    <button type="button" aria-label="Предыдущий месяц" disabled={monthKey <= currentMonthKey} onClick={() => changeMonth(-1)}>‹</button>
                    <h3>{calendarMonth.toLocaleDateString("ru-RU", { month: "long", year: "numeric" })}</h3>
                    <button type="button" aria-label="Следующий месяц" onClick={() => changeMonth(1)}>›</button>
                  </div>

                  <div className="date-calendar-grid" role="group" aria-label="Выберите одну или несколько дат">
                    {["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"].map((day) => (
                      <span className="date-calendar-weekday" key={day}>{day}</span>
                    ))}
                    {Array.from({ length: firstWeekday }, (_, index) => (
                      <span className="date-calendar-empty" key={`empty-${index}`} aria-hidden="true" />
                    ))}
                    {Array.from({ length: daysInMonth }, (_, index) => {
                      const dayNumber = index + 1;
                      const dateKey = `${monthKey}-${String(dayNumber).padStart(2, "0")}`;
                      const isPast = dateKey < todayKey;
                      const isSelected = selectedDates.includes(dateKey);
                      return (
                        <button
                          type="button"
                          key={dateKey}
                          className={`date-calendar-day${isSelected ? " selected" : ""}`}
                          disabled={isPast}
                          aria-pressed={isSelected}
                          aria-label={new Date(calendarYear, calendarMonthIndex, dayNumber).toLocaleDateString("ru-RU", { day: "numeric", month: "long", year: "numeric" })}
                          onClick={() => toggleDate(dateKey)}
                        >
                          {dayNumber}
                        </button>
                      );
                    })}
                  </div>

                  <div className="time-filter">
                    <h3>Время</h3>
                    <div className="time-presets" aria-label="Время суток">
                      {[
                        ["any", "Любое"],
                        ["morning", "Утро"],
                        ["day", "День"],
                        ["evening", "Вечер"],
                      ].map(([value, label]) => (
                        <button
                          type="button"
                          key={value}
                          className={timeOfDay === value ? "selected" : ""}
                          aria-pressed={timeOfDay === value}
                          onClick={() => setTimeOfDay(value)}
                        >
                          {label}
                        </button>
                      ))}
                    </div>
                    <label className="exact-time-label" htmlFor="exact-time">Точное время</label>
                    <input id="exact-time" className="exact-time-input" type="time" value={exactTime} onChange={(event) => setExactTime(event.target.value)} />
                  </div>

                  <div className="date-actions">
                    <button type="button" className="reset-date" onClick={resetDateFilter}>Сбросить</button>
                    <button type="button" className="done-date" onClick={() => {
                      setAppliedDateFilter({
                        selectedDates: [...selectedDates],
                        datePreset,
                        timeOfDay,
                        exactTime,
                      });
                      setDateOpen(false);
                    }}>Готово</button>
                  </div>
                </section>
              )}

              {cityOpen && (
                <div className="city-dropdown">
                  <div className="country-switch">
                    {["Беларусь", "Россия"].map((item) => (
                      <button
                        key={item}
                        className={country === item ? "selected" : ""}
                        onClick={() => {
                          setCountry(item);
                          setCity("");
                          setCitySearch("");
                        }}
                      >
                        {item}
                      </button>
                    ))}
                  </div>

                  <input
                    className="city-search"
                    placeholder="Найти город"
                    value={citySearch}
                    onChange={(e) => setCitySearch(e.target.value)}
                  />

                  <p className="dropdown-label">Популярные</p>
                  <div className="city-chips">
                    {cities[country].popular
                      .filter((item) => item.toLowerCase().includes(citySearch.toLowerCase()))
                      .map((item) => (
                        <button
                          key={item}
                          className={city === item ? "selected" : ""}
                          onClick={() => {
                            setCity(item);
                            setCityOpen(false);
                          }}
                        >
                          {item}
                        </button>
                      ))}
                  </div>

                  {nearbyCities.length > 0 && !citySearch && (
                    <>
                      <p className="dropdown-label">Города-спутники</p>
                      <div className="city-chips">
                        {nearbyCities.map((item) => (
                          <button key={item} onClick={() => {
                            setCity(item);
                            setCityOpen(false);
                          }}>
                            {item}
                          </button>
                        ))}
                      </div>
                    </>
                  )}

                  <button className="clear-city" onClick={() => {
                    setCity("");
                    setCityOpen(false);
                  }}>
                    Все города
                  </button>
                </div>
              )}
            </div>
          </div>

          <div className="results-heading">
            <h2>Мастера</h2>
            <span>{filteredMasters.length}</span>
          </div>

          <div className="master-list">
            {filteredMasters.map((master) => (
              <article
                className="master-card"
                key={master.id}
                onClick={() => setSelectedMaster(master)}
              >
                <div className="master-image">
                  <img src={master.image} alt={master.name} />
                  <span className="master-rating">★ {master.rating.toFixed(1)}</span>
                </div>

                <div className="master-card-info">
                  <h3>{master.name}</h3>
                  <div className="service-tags">
                    {master.services.map((service) => (
                      <span key={service}>{service}</span>
                    ))}
                  </div>
                  <div className="master-card-bottom">
                    <span>⌖ {master.city}</span>
                    <strong>от {master.price} BYN</strong>
                  </div>
                </div>
              </article>
            ))}

            {filteredMasters.length === 0 && (
              <div className="no-results">
                Мастера не найдены. Попробуйте изменить запрос.
              </div>
            )}
          </div>
        </section>
      )}

      {tab === "search" && selectedMaster && (
        <section className="master-detail">
          <button className="back-link" onClick={() => setSelectedMaster(null)}>
            ← К поиску
          </button>
          <img src={selectedMaster.image} alt={selectedMaster.name} />
          <h1>{selectedMaster.name}</h1>
          <p>{selectedMaster.city}</p>
          <p>Услуги: {selectedMaster.services.join(", ")}</p>
          <p>Стоимость от {selectedMaster.price} BYN</p>
          <button className="find-button booking-action" onClick={() => alert("Выбор услуги и времени добавим следующим этапом")}>
            Записаться
          </button>
        </section>
      )}

      {tab === "bookings" && (
        <section>
          <h1 className="client-title">Мои записи</h1>
          <div className="booking-tabs">
            {[
              ["upcoming", "Предстоящие"],
              ["history", "История"],
              ["cancelled", "Отменённые"],
            ].map(([value, label]) => (
              <button
                key={value}
                className={bookingTab === value ? "active" : ""}
                onClick={() => setBookingTab(value)}
              >
                {label}
              </button>
            ))}
          </div>
          <div className="no-results">Здесь будут ваши записи</div>
        </section>
      )}

      {tab === "profile" && (
        <section>
          <div className="profile-heading">
            <h1 className="client-title">Профиль</h1>
            <button className="profile-settings">⚙</button>
          </div>
          <div className="client-profile-card">
            <div className="client-profile-photo">А</div>
            <h2>Александр</h2>
            <p>@telegram</p>
            <button>Редактировать профиль</button>
          </div>
        </section>
      )}

      <nav className="client-bottom-nav">
        <button className={tab === "search" ? "active" : ""} onClick={() => {
          setTab("search");
          setSelectedMaster(null);
        }}>
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

function toDateKey(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatDateLabel(dateKey) {
  const [year, month, day] = dateKey.split("-").map(Number);
  return new Date(year, month - 1, day).toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}

function formatDateFilterLabel({ selectedDates, datePreset, timeOfDay, exactTime }) {
  const labels = [];
  if (datePreset === "today") labels.push("Сегодня");
  else if (datePreset === "tomorrow") labels.push("Завтра");
  else if (datePreset === "weekends") labels.push("Выходные");
  else if (selectedDates.length === 1) labels.push(formatDateLabel(selectedDates[0]));
  else if (selectedDates.length > 1) labels.push(`${selectedDates.length} ${dateCountWord(selectedDates.length)}`);

  if (timeOfDay !== "any") {
    labels.push({ morning: "Утро", day: "День", evening: "Вечер" }[timeOfDay]);
  }
  if (exactTime) labels.push(exactTime);
  return labels.length ? labels.join(" · ") : "Когда?";
}

function dateCountWord(count) {
  const lastTwoDigits = count % 100;
  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return "дат";
  const lastDigit = count % 10;
  if (lastDigit === 1) return "дата";
  if (lastDigit >= 2 && lastDigit <= 4) return "даты";
  return "дат";
}

export default ClientCabinet;
