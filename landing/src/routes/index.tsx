import { createFileRoute } from "@tanstack/react-router";
import {
  Globe,
  Send,
  Activity,
  Wifi,
  Wallet,
  Warehouse,
  Gauge,
  FlaskConical,
  Wind,
  Sparkles,
  MapPin,
  Phone,
  Clock,
  QrCode,
  CarFront,
  ShieldCheck,
  CreditCard,
  RotateCcw,
  UserRound,
  ChevronRight,
  Droplets,
} from "lucide-react";

import logoUrl from "../assets/h2o-logo.webp";

// TODO: Подставь реальный URL мини-приложения (веб-версии).
const WEB_APP_URL = "https://h2o-nsk.ru/web/";
const TELEGRAM_BOT_URL = "https://t.me/h2o_nsk_bot";
const STATUS_URL = "https://h2o-nsk.ru/status";
const PHONE = "+7 (383) 287-03-78";
const ADDRESS = 'г. Новосибирск, ул. Доватора, 11 (ТЦ «Автоградъ»)';

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "H2O — автомойка самообслуживания в Новосибирске" },
      {
        name: "description",
        content:
          "Тёплая автомойка самообслуживания H2O: 20 постов, 200 бар, оплата с сайта или Telegram-бота. Доватора, 11. Работаем 24/7.",
      },
      { property: "og:title", content: "H2O — автомойка самообслуживания" },
      {
        property: "og:description",
        content:
          "20 тёплых постов, итальянское оборудование, оплата онлайн. Доватора, 11, Новосибирск. 24/7.",
      },
    ],
  }),
  component: Index,
});

function Index() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <Header />
      <main>
        <Hero />
        <Features />
        <Pricing />
        <HowTo />
        <Support />
        <Contacts />
        <CTASection />
      </main>
      <Footer />
    </div>
  );
}

/* ---------------- Header ---------------- */

function Header() {
  return (
    <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur-xl">
      <div className="mx-auto flex h-28 max-w-6xl items-center justify-between px-5">
        <a href="#top" className="flex items-center gap-2">
          <img
            src={logoUrl}
            alt="H2O — автомойка самообслуживания"
            className="h-24 w-auto"
          />
        </a>
        <nav className="hidden items-center gap-7 text-sm text-muted-foreground lg:flex">
          <a href="#features" className="transition-colors hover:text-foreground">Преимущества</a>
          <a href="#pricing" className="transition-colors hover:text-foreground">Цены</a>
          <a href="#howto" className="transition-colors hover:text-foreground">Как мыть</a>
          <a href="#contacts" className="transition-colors hover:text-foreground">Контакты</a>
        </nav>
        <a
          href={WEB_APP_URL}
          target="_blank"
          rel="noopener noreferrer"
          className="hidden items-center gap-1.5 rounded-full bg-brand px-5 py-2.5 text-sm font-semibold text-brand-foreground shadow-glow transition-transform hover:scale-[1.03] sm:inline-flex"
        >
          Помыть машину
          <ChevronRight className="h-4 w-4" />
        </a>
      </div>
    </header>
  );
}

/* ---------------- Hero ---------------- */

function Hero() {
  return (
    <section id="top" className="relative overflow-hidden bg-hero">
      <div className="pointer-events-none absolute inset-0 opacity-[0.07] [background-image:radial-gradient(circle_at_1px_1px,white_1px,transparent_0)] [background-size:24px_24px]" />
      <div className="relative mx-auto max-w-6xl px-5 pb-24 pt-20 md:pt-28">
        <div className="max-w-3xl">
          <span className="inline-flex items-center gap-2 rounded-full border border-border bg-surface/60 px-3 py-1 text-xs font-medium text-muted-foreground backdrop-blur">
            <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-brand" />
            Работаем 24/7 с 28 апреля 2009 года
          </span>
          <h1 className="mt-6 font-display text-4xl font-extrabold leading-[1.05] tracking-tight md:text-6xl">
            Автомойка самообслуживания{" "}
            <span className="text-brand font-extrabold">H2O</span> в Новосибирске
          </h1>
          <p className="mt-5 max-w-2xl text-lg text-muted-foreground md:text-xl">
            Мой хорошо — мой сам. 20 тёплых постов, итальянское оборудование 200 бар и оплата
            прямо с телефона — на сайте или в Telegram-боте.
          </p>

          <p className="mt-5 inline-flex items-center gap-2 rounded-xl border border-brand/40 bg-brand/10 px-4 py-2.5 text-sm font-medium text-foreground md:text-base">
            <Wifi className="h-5 w-5 shrink-0 text-brand" />
            Бесплатный Wi-Fi на мойке — оплатите через приложение, даже если мобильный интернет не ловит.
          </p>

          <div className="mt-9 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <CTA href={STATUS_URL} icon={<Activity className="h-5 w-5" />} ghost>
              Статус загруженности онлайн
            </CTA>
            <CTA href={WEB_APP_URL} icon={<Globe className="h-5 w-5" />} primary>
              Оплатить и управлять мойкой на сайте
            </CTA>
            <CTA href={TELEGRAM_BOT_URL} icon={<Send className="h-5 w-5" />}>
              Управлять мойкой в Telegram боте
            </CTA>
          </div>

          <div className="mt-10 flex flex-wrap items-center gap-x-8 gap-y-4 text-sm text-muted-foreground">
            <Stat value="20" label="постов" />
            <Stat value="200 бар" label="давление" />
            <Stat value="от 3.5 ₽/мин" label="оплата по факту" />
            <Stat value="24/7" label="режим работы" />
          </div>
        </div>
      </div>
    </section>
  );
}

function CTA({
  href,
  icon,
  children,
  primary,
  ghost,
}: {
  href: string;
  icon: React.ReactNode;
  children: React.ReactNode;
  primary?: boolean;
  ghost?: boolean;
}) {
  const base =
    "inline-flex items-center justify-center gap-2 rounded-xl px-5 py-3 text-base font-semibold transition-all duration-200";
  const style = primary
    ? "bg-brand text-brand-foreground hover:scale-[1.02] shadow-glow"
    : ghost
      ? "bg-brand-gradient text-brand-foreground hover:scale-[1.02] shadow-glow"
      : "bg-brand text-brand-foreground hover:scale-[1.02] shadow-glow";
  return (
    <a href={href} target="_blank" rel="noopener noreferrer" className={`${base} ${style}`}>
      {icon}
      {children}
    </a>
  );
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div>
      <div className="font-display text-2xl font-bold text-foreground">{value}</div>
      <div className="text-xs uppercase tracking-wider">{label}</div>
    </div>
  );
}

/* ---------------- Features ---------------- */

const FEATURES: {
  icon: typeof Wifi;
  title: string;
  text: string;
  highlight?: boolean;
}[] = [
  {
    icon: Wifi,
    title: "Бесплатный Wi-Fi",
    text: "Бесплатный Wi-Fi на всей мойке — спокойно оплатите через приложение, даже если мобильный интернет барахлит.",
    highlight: true,
  },
  {
    icon: Wallet,
    title: "В 10 раз дешевле",
    text: "В 10 раз дешевле любой мойки самообслуживания. Всего 10 ₽/мин.",
  },
  {
    icon: Warehouse,
    title: "20 тёплых постов",
    text: "Тёплое помещение и большая площадка — очередь движется быстро.",
  },
  {
    icon: Gauge,
    title: "Мощное оборудование",
    text: "Итальянские аппараты высокого давления 200 бар, 1000 л/час.",
  },
  {
    icon: FlaskConical,
    title: "Качественная химия",
    text: "Сильная химия для мойки и средства для чистки и ухода за салоном.",
  },
  {
    icon: Wind,
    title: "Мощные пылесосы",
    text: "Трёхтурбинные пылесосы — справляются даже с водой в салоне.",
  },
  {
    icon: Sparkles,
    title: "Продувка воздухом",
    text: "Воздух под давлением для продувки замков, щелей и зеркал.",
  },
  {
    icon: CarFront,
    title: "Просто и понятно",
    text: "Интерфейс настолько простой, что справляется каждый.",
  },
  {
    icon: MapPin,
    title: "Удобное расположение",
    text: "Кольцо в конце улицы Бориса Богаткова, ТЦ «Автоградъ».",
  },
];

function Features() {
  return (
    <section id="features" className="mx-auto max-w-6xl px-5 py-24">
      <SectionHeader
        eyebrow="Почему H2O"
        title="Всё для того, чтобы мыть машину было удобно"
      />
      <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {FEATURES.map(({ icon: Icon, title, text, highlight }) => (
          <div
            key={title}
            className={
              highlight
                ? "group relative rounded-2xl border-2 border-brand bg-brand/[0.07] p-6 shadow-glow"
                : "group rounded-2xl border border-border bg-surface/60 p-6 transition-colors hover:border-brand/40"
            }
          >
            {highlight && (
              <span className="absolute right-4 top-4 rounded-full bg-brand px-2.5 py-1 text-[11px] font-bold uppercase tracking-wider text-brand-foreground">
                Важно сейчас
              </span>
            )}
            <div
              className={
                highlight
                  ? "grid h-11 w-11 place-items-center rounded-xl bg-brand text-brand-foreground"
                  : "grid h-11 w-11 place-items-center rounded-xl bg-brand/10 text-brand transition-colors group-hover:bg-brand group-hover:text-brand-foreground"
              }
            >
              <Icon className="h-5 w-5" />
            </div>
            <h3 className="mt-5 font-display text-lg font-bold">{title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{text}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* ---------------- Pricing ---------------- */

const PRICES: { name: string; price: string }[] = [
  { name: "Аппарат высокого давления", price: "10" },
  { name: "Распылитель химии", price: "20" },
  { name: "Пылесос", price: "8" },
  { name: "Продувочный пистолет", price: "3.5" },
];


function Pricing() {
  return (
    <section id="pricing" className="border-y border-border bg-surface/40">
      <div className="mx-auto max-w-6xl px-5 py-24">
        <SectionHeader
          eyebrow="Цены"
          title="Платите только за то, чем пользуетесь"
        />
        <div className="mt-12 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {PRICES.map((p) => (
            <div
              key={p.name}
              className="relative rounded-2xl border border-border bg-background p-6"
            >
              <div className="text-sm text-muted-foreground">{p.name}</div>
              <div className="mt-4 flex items-baseline gap-1">
                <span className="font-display text-4xl font-extrabold text-foreground">
                  {p.price}
                </span>
                <span className="text-base text-muted-foreground">₽/мин</span>
              </div>
            </div>
          ))}
        </div>
        <p className="mt-6 text-xs text-muted-foreground">
          * Минимально допустимое время варьируется в зависимости от ваших потребностей.
        </p>
      </div>
    </section>
  );
}

/* ---------------- How to use ---------------- */

const STEPS = [
  {
    icon: CarFront,
    title: "Приезжайте",
    text: "Приезжайте к нам на автомойку в любое удобное время — мы работаем круглосуточно.",
  },
  {
    icon: QrCode,
    title: "Откройте веб-версию или бота",
    text: (
      <>
        Отсканируйте QR-код на стене и откройте мини-приложение —{" "}
        <a href={WEB_APP_URL} className="text-brand underline-offset-4 hover:underline" target="_blank" rel="noopener noreferrer">
          на сайте
        </a>{" "}
        или в{" "}
        <a href={TELEGRAM_BOT_URL} className="text-brand underline-offset-4 hover:underline" target="_blank" rel="noopener noreferrer">
          Telegram-боте
        </a>
        . Можно и просто нажать кнопку «Помыть машину» здесь же.
      </>
    ),
  },
  {
    icon: CreditCard,
    title: "Выберите услуги и оплатите",
    text: "Выберите нужные услуги в приложении и оплатите — картой или через СБП.",
  },
  {
    icon: Activity,
    title: "Дождитесь назначения бокса",
    text: "Дождитесь назначения свободного бокса — придёт уведомление, номер также появится на экране.",
  },
  {
    icon: Droplets,
    title: "Заезжайте и мойте",
    text: "Заезжайте в назначенный бокс, включите его в приложении и наслаждайтесь. Можно продлить или закончить раньше.",
  },
] as const;

function HowTo() {
  return (
    <section id="howto" className="mx-auto max-w-6xl px-5 py-24">
      <SectionHeader eyebrow="Как пользоваться" title="Пять простых шагов" />

      <ol className="mt-12 grid grid-cols-1 gap-4 md:grid-cols-2">
        {STEPS.map((s, i) => (
          <li
            key={s.title}
            className="relative rounded-2xl border border-border bg-surface/60 p-6"
          >
            <div className="flex items-start gap-4">
              <div className="grid h-12 w-12 shrink-0 place-items-center rounded-xl bg-brand-gradient text-brand-foreground shadow-glow">
                <s.icon className="h-6 w-6" />
              </div>
              <div>
                <div className="text-xs font-semibold uppercase tracking-wider text-brand">
                  Шаг {i + 1}
                </div>
                <h3 className="mt-1 font-display text-lg font-bold">{s.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{s.text}</p>
              </div>
            </div>
          </li>
        ))}
      </ol>

      <div className="mt-10 flex flex-wrap items-center justify-center gap-3 rounded-2xl border border-border bg-surface/40 p-6">
        <span className="mr-2 text-sm text-muted-foreground">Принимаем оплату:</span>
        {["Т-Банк", "СБП", "Карты", "Сбербанк", "QR-код", "Наличные"].map((method) => (
          <span key={method} className="rounded-full border border-border bg-surface px-3 py-1 text-sm text-foreground">
            {method}
          </span>
        ))}
      </div>
    </section>
  );
}

/* ---------------- Support ---------------- */

const SUPPORT = [
  {
    icon: UserRound,
    title: "Нужна помощь?",
    text: "Если у вас проблемы с интернетом, Telegram или только наличные — всё то же самое для вас сделает наш администратор-кассир.",
  },
  {
    icon: ShieldCheck,
    title: "Гарантия обслуживания",
    text: "В случае поломки во время оплаченной мойки — переставим в другой бокс без очереди и включим ваше время. Обращайтесь к кассиру-администратору.",
  },
  {
    icon: RotateCcw,
    title: "Возврат средств",
    text: "Возврат средств предусмотрен: в приложении (до старта мойки) или у кассира-администратора (после).",
  },
] as const;

function Support() {
  return (
    <section className="border-y border-border bg-surface/40">
      <div className="mx-auto max-w-6xl px-5 py-24">
        <SectionHeader eyebrow="Поддержка" title="Мы всегда рядом" />
        <div className="mt-12 grid grid-cols-1 gap-4 md:grid-cols-3">
          {SUPPORT.map((s) => (
            <div key={s.title} className="rounded-2xl border border-border bg-background p-6">
              <div className="grid h-11 w-11 place-items-center rounded-xl bg-brand/10 text-brand">
                <s.icon className="h-5 w-5" />
              </div>
              <h3 className="mt-5 font-display text-lg font-bold">{s.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{s.text}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ---------------- Contacts ---------------- */

function Contacts() {
  return (
    <section id="contacts" className="mx-auto max-w-6xl px-5 py-24">
      <SectionHeader eyebrow="Контакты" title="Как нас найти" />
      <div className="mt-12 grid grid-cols-1 gap-4 md:grid-cols-3">
        <ContactCard icon={MapPin} title="Адрес" lines={[ADDRESS, "Кольцо в конце ул. Бориса Богаткова"]} />
        <ContactCard
          icon={Phone}
          title="Телефон"
          lines={[
            <a key="p" href={`tel:${PHONE.replace(/[^+\d]/g, "")}`} className="text-foreground hover:text-brand">
              {PHONE}
            </a>,
          ]}
        />
        <ContactCard icon={Clock} title="Режим работы" lines={["Круглосуточно, 24/7", "Без выходных"]} />
      </div>
    </section>
  );
}

function ContactCard({
  icon: Icon,
  title,
  lines,
}: {
  icon: typeof MapPin;
  title: string;
  lines: React.ReactNode[];
}) {
  return (
    <div className="rounded-2xl border border-border bg-surface/60 p-6">
      <div className="flex items-center gap-3">
        <div className="grid h-10 w-10 place-items-center rounded-xl bg-brand/10 text-brand">
          <Icon className="h-5 w-5" />
        </div>
        <div className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {title}
        </div>
      </div>
      <div className="mt-4 space-y-1 text-base text-foreground">
        {lines.map((l, i) => (
          <div key={i}>{l}</div>
        ))}
      </div>
    </div>
  );
}

/* ---------------- CTA Section ---------------- */

function CTASection() {
  return (
    <section className="mx-auto max-w-6xl px-5 pb-24">
      <div className="relative overflow-hidden rounded-3xl border border-border bg-hero p-10 md:p-16">
        <div className="relative max-w-2xl">
          <h2 className="font-display text-3xl font-extrabold tracking-tight md:text-4xl">
            Начните мойку прямо сейчас
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            Откройте веб-версию или Telegram-бот, выберите услуги и приезжайте к нам.
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <CTA href={WEB_APP_URL} icon={<Globe className="h-5 w-5" />} primary>
              Помыть на сайте
            </CTA>
            <CTA href={TELEGRAM_BOT_URL} icon={<Send className="h-5 w-5" />}>
              Открыть Telegram-бот
            </CTA>
          </div>
        </div>
      </div>
    </section>
  );
}

/* ---------------- Footer ---------------- */

function Footer() {
  return (
    <footer className="border-t border-border bg-surface/40">
      <div className="mx-auto flex max-w-6xl flex-col gap-6 px-5 py-10 md:flex-row md:items-center md:justify-between">
        <div className="flex items-center gap-3">
          <img
            src={logoUrl}
            alt="H2O"
            className="h-16 w-auto"
          />
          <span className="text-sm text-muted-foreground">Автомойка самообслуживания, Новосибирск</span>
        </div>
        <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-muted-foreground">
          <a href="/oferta" className="hover:text-foreground">Публичная оферта</a>
          <a href={WEB_APP_URL} target="_blank" rel="noopener noreferrer" className="hover:text-foreground">Веб-версия</a>
          <a href={TELEGRAM_BOT_URL} target="_blank" rel="noopener noreferrer" className="hover:text-foreground">Telegram-бот</a>
          <span>© {new Date().getFullYear()} H2O</span>
        </div>
      </div>
    </footer>
  );
}

/* ---------------- Section header ---------------- */

function SectionHeader({ eyebrow, title }: { eyebrow: string; title: string }) {
  return (
    <div className="max-w-2xl">
      <div className="text-xs font-semibold uppercase tracking-[0.18em] text-brand">{eyebrow}</div>
      <h2 className="mt-3 font-display text-3xl font-extrabold tracking-tight md:text-4xl">
        {title}
      </h2>
    </div>
  );
}
