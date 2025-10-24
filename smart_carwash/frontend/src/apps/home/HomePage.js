import React from 'react';
import styled from 'styled-components';

const HomeContainer = styled.div`
  min-height: 100vh;
  background: url('/images/water-drops-bg.jpg') center center;
  background-size: cover;
  background-attachment: fixed;
  color: white;
  position: relative;
  overflow-x: hidden;
  
  @media (max-width: 768px) {
    background-attachment: scroll;
  }
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, rgba(74, 144, 226, 0.7) 0%, rgba(91, 163, 208, 0.7) 100%);
    pointer-events: none;
    z-index: 1;
  }
  
  & > * {
    position: relative;
    z-index: 2;
  }
`;

const Header = styled.header`
  background: rgba(0, 0, 0, 0.2);
  padding: 20px;
  text-align: center;
  backdrop-filter: blur(10px);
`;

const Logo = styled.h1`
  font-size: 3rem;
  margin: 0;
  font-weight: bold;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
  
  @media (max-width: 768px) {
    font-size: 2rem;
  }
`;

const Tagline = styled.p`
  font-size: 1.2rem;
  margin: 10px 0 0 0;
  opacity: 0.9;
  
  @media (max-width: 768px) {
    font-size: 1rem;
  }
`;

const TelegramButton = styled.a`
  display: inline-block;
  margin-top: 20px;
  padding: 12px 30px;
  background: linear-gradient(135deg, #0088cc 0%, #0077b5 100%);
  color: white;
  text-decoration: none;
  border-radius: 25px;
  font-weight: bold;
  font-size: 1.1rem;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(0, 136, 204, 0.3);
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(0, 136, 204, 0.5);
    background: linear-gradient(135deg, #0099dd 0%, #0088cc 100%);
    text-decoration: none;
  }
  
  @media (max-width: 768px) {
    font-size: 1rem;
    padding: 10px 25px;
  }
`;

const MainContent = styled.main`
  max-width: 1200px;
  margin: 0 auto;
  padding: 40px 20px;
  
  @media (max-width: 768px) {
    padding: 20px 10px;
  }
`;

const Section = styled.section`
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 30px;
  margin-bottom: 30px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
  
  @media (max-width: 768px) {
    padding: 20px 15px;
    margin-bottom: 20px;
  }
`;

const SectionTitle = styled.h2`
  font-size: 2rem;
  margin-bottom: 20px;
  text-align: center;
  
  @media (max-width: 768px) {
    font-size: 1.5rem;
  }
`;

const FeatureGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-top: 30px;
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    gap: 15px;
    margin-top: 20px;
  }
`;

const FeatureCard = styled.div`
  background: rgba(255, 255, 255, 0.07);
  padding: 25px;
  border-radius: 15px;
  text-align: center;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  
  &:hover {
    transform: translateY(-5px);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
    background: rgba(255, 255, 255, 0.12);
  }
  
  @media (max-width: 768px) {
    padding: 20px 15px;
  }
`;

const FeatureIcon = styled.div`
  font-size: 3rem;
  margin-bottom: 15px;
`;

const FeatureTitle = styled.h3`
  font-size: 1.3rem;
  margin-bottom: 10px;
`;

const FeatureDescription = styled.p`
  font-size: 1rem;
  opacity: 0.9;
  line-height: 1.5;
`;

const PriceList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 20px;
`;

const PriceItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.07);
  padding: 20px;
  border-radius: 10px;
  
  @media (max-width: 768px) {
    flex-direction: column;
    text-align: center;
    gap: 10px;
  }
`;

const ServiceName = styled.span`
  font-size: 1.2rem;
  font-weight: 500;
`;

const Price = styled.span`
  font-size: 1.5rem;
  font-weight: bold;
  color: #FFE066;
`;

const ContactInfo = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-top: 20px;
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    gap: 15px;
  }
`;

const ContactItem = styled.div`
  background: rgba(255, 255, 255, 0.07);
  padding: 20px;
  border-radius: 10px;
  text-align: center;
`;

const ContactLabel = styled.div`
  font-size: 1rem;
  opacity: 0.8;
  margin-bottom: 5px;
`;

const ContactValue = styled.div`
  font-size: 1.2rem;
  font-weight: bold;
  
  @media (max-width: 768px) {
    font-size: 1rem;
  }
`;

const CTASection = styled.section`
  background: linear-gradient(135deg, rgba(0, 136, 204, 0.2) 0%, rgba(0, 119, 181, 0.2) 100%);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  padding: 50px 30px;
  margin-bottom: 30px;
  text-align: center;
  border: 2px solid rgba(0, 136, 204, 0.3);
  
  @media (max-width: 768px) {
    padding: 30px 20px;
  }
`;

const CTATitle = styled.h2`
  font-size: 2.5rem;
  margin-bottom: 15px;
  
  @media (max-width: 768px) {
    font-size: 1.8rem;
  }
`;

const CTADescription = styled.p`
  font-size: 1.2rem;
  margin-bottom: 30px;
  opacity: 0.9;
  
  @media (max-width: 768px) {
    font-size: 1rem;
  }
`;

const TelegramLink = styled.a`
  color: #FFE066;
  text-decoration: underline;
  font-weight: bold;
  transition: opacity 0.3s ease;
  
  &:hover {
    opacity: 0.8;
    color: #FFE066;
  }
`;

const PaymentLogos = styled.div`
  display: flex;
  gap: 15px;
  justify-content: center;
  align-items: center;
  margin-top: 15px;
  flex-wrap: wrap;
`;

const PaymentLogo = styled.img`
  height: 40px;
  width: auto;
  object-fit: contain;
  background: white;
  padding: 8px;
  border-radius: 8px;
  
  @media (max-width: 768px) {
    height: 35px;
  }
`;

const Footer = styled.footer`
  background: rgba(0, 0, 0, 0.2);
  padding: 20px;
  text-align: center;
  margin-top: 40px;
`;

const HomePage = () => {
  return (
    <HomeContainer>
      <Header>
        <Logo>H2O - Автомойка Самообслуживания</Logo>
        <Tagline>Мой хорошо - мой сам! С 28 апреля 2009 года</Tagline>
        <TelegramButton href="https://t.me/h2o_nsk_bot" target="_blank" rel="noopener noreferrer">
          📱 Открыть Telegram бот
        </TelegramButton>
      </Header>

      <MainContent>
        <Section>
          <SectionTitle>🌟 Наши Преимущества</SectionTitle>
          <FeatureGrid>
            <FeatureCard>
              <FeatureIcon>💰</FeatureIcon>
              <FeatureTitle>В 10 раз дешевле!</FeatureTitle>
              <FeatureDescription>
                В 10 раз дешевле любой мойки самообслуживания. Всего 10 рублей/минута!
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🚗</FeatureIcon>
              <FeatureTitle>Большая автомойка</FeatureTitle>
              <FeatureDescription>
                Теплое помещение, 20 постов. Очень быстрая очередь
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>💪</FeatureIcon>
              <FeatureTitle>Мощное оборудование</FeatureTitle>
              <FeatureDescription>
                Итальянские аппараты высокого давления 200 бар, 1000 литров в час. Почувствуй мощь!
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🧪</FeatureIcon>
              <FeatureTitle>Качественная химия</FeatureTitle>
              <FeatureDescription>
                Качественная и сильная химия для мойки. Химия для чистки и ухода за салоном
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🌪️</FeatureIcon>
              <FeatureTitle>Мощные пылесосы</FeatureTitle>
              <FeatureDescription>
                Трехтурбинные пылесосы (можно пылесосить воду)
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🌬️</FeatureIcon>
              <FeatureTitle>Продувка</FeatureTitle>
              <FeatureDescription>
                Воздух для продувки замков и щелей
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🎯</FeatureIcon>
              <FeatureTitle>Просто и понятно</FeatureTitle>
              <FeatureDescription>
                Интуитивно понятно и интересно. Справляются все!
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>📍</FeatureIcon>
              <FeatureTitle>Удобное расположение</FeatureTitle>
              <FeatureDescription>
                Кольцо в конце улицы Бориса Богаткова
              </FeatureDescription>
            </FeatureCard>
          </FeatureGrid>
        </Section>

        <Section>
          <SectionTitle>💳 Цены</SectionTitle>
          <PriceList>
            <PriceItem>
              <ServiceName>Аппарат высокого давления</ServiceName>
              <Price>10 ₽/мин</Price>
            </PriceItem>
            <PriceItem>
              <ServiceName>Хороший распылитель химии</ServiceName>
              <Price>20 ₽/мин</Price>
            </PriceItem>
            <PriceItem>
              <ServiceName>Пылесос</ServiceName>
              <Price>8 ₽/мин</Price>
            </PriceItem>
            <PriceItem>
              <ServiceName>Продувочный пистолет</ServiceName>
              <Price>3.5 ₽/мин</Price>
            </PriceItem>
          </PriceList>
          <FeatureDescription style={{marginTop: '20px', textAlign: 'center', fontSize: '0.95rem'}}>
            * Минимально допустимое время варьируется в зависимости от ваших потребностей
          </FeatureDescription>
        </Section>

        <Section>
          <SectionTitle>📍 Контакты и Адрес</SectionTitle>
          <ContactInfo>
            <ContactItem>
              <ContactLabel>Адрес</ContactLabel>
              <ContactValue>г. Новосибирск, ул. Доватора, д. 11<br/>(здание ТЦ "Автоградъ")</ContactValue>
            </ContactItem>
            
            <ContactItem>
              <ContactLabel>Телефон</ContactLabel>
              <ContactValue>+7 (383) 287-03-78</ContactValue>
            </ContactItem>
            
            <ContactItem>
              <ContactLabel>Режим работы</ContactLabel>
              <ContactValue>Круглосуточно, 24/7</ContactValue>
            </ContactItem>
          </ContactInfo>
        </Section>

        <Section>
          <SectionTitle>🎓 Как пользоваться?</SectionTitle>
          <FeatureGrid>
            <FeatureCard>
              <FeatureIcon>1️⃣</FeatureIcon>
              <FeatureTitle>Приезжайте</FeatureTitle>
              <FeatureDescription>
                Приезжайте к нам на автомойку
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>2️⃣</FeatureIcon>
              <FeatureTitle>Сканируйте QR-код</FeatureTitle>
              <FeatureDescription>
                Отсканируйте QR-код и <TelegramLink href="https://t.me/h2o_nsk_bot" target="_blank" rel="noopener noreferrer">откройте мини-приложение в Telegram-боте</TelegramLink>
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>3️⃣</FeatureIcon>
              <FeatureTitle>Выберите услуги</FeatureTitle>
              <FeatureDescription>
                Выберите нужные услуги и оплатите в приложении
              </FeatureDescription>
              <PaymentLogos>
                <PaymentLogo src="/images/tbank-logo.png" alt="Т-Банк" />
                <PaymentLogo src="/images/sbp-logo.png" alt="Система быстрых платежей" />
              </PaymentLogos>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>4️⃣</FeatureIcon>
              <FeatureTitle>Дождитесь назначения</FeatureTitle>
              <FeatureDescription>
                Дождитесь назначения свободного бокса (придет уведомление в Telegram)
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>5️⃣</FeatureIcon>
              <FeatureTitle>Заезжайте и мойте!</FeatureTitle>
              <FeatureDescription>
                Смело заезжайте в назначенный бокс, включите его в приложении и наслаждайтесь процессом. Можно продлить или закончить раньше
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>👤</FeatureIcon>
              <FeatureTitle>Нужна помощь?</FeatureTitle>
              <FeatureDescription>
                Если у вас проблемы с интернетом, Telegram или только <strong style={{color: '#FFE066', fontSize: '1.1em'}}>наличные</strong> - всё то же самое может для вас сделать наш администратор-кассир
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>🛡️</FeatureIcon>
              <FeatureTitle>Гарантия обслуживания</FeatureTitle>
              <FeatureDescription>
                В случае поломки во время оплаченной мойки - мы переставим вас в другой бокс без очереди и включим ваше время. Обращайтесь к кассиру-администратору
              </FeatureDescription>
            </FeatureCard>
            
            <FeatureCard>
              <FeatureIcon>💰</FeatureIcon>
              <FeatureTitle>Возврат средств</FeatureTitle>
              <FeatureDescription>
                Возврат средств предусмотрен: в приложении (до старта мойки) или у кассира-администратора (после)
              </FeatureDescription>
            </FeatureCard>
          </FeatureGrid>
        </Section>

        <CTASection>
          <CTATitle>🚀 Начните мойку прямо сейчас!</CTATitle>
          <CTADescription>
            Откройте наш Telegram-бот, выберите услуги и приезжайте на мойку
          </CTADescription>
          <TelegramButton href="https://t.me/h2o_nsk_bot" target="_blank" rel="noopener noreferrer">
            📱 Открыть @h2o_nsk_bot
          </TelegramButton>
        </CTASection>
      </MainContent>

      <Footer>
        <p>© 2009-2025 H2O Автомойка Самообслуживания. Все права защищены.</p>
        <p style={{marginTop: '10px', fontSize: '0.9rem', opacity: '0.8'}}>С вами с 28 апреля 2009 года</p>
        <p style={{marginTop: '15px', fontSize: '0.85rem', opacity: '0.7'}}>
          ООО "Эталон"<br/>
          ИНН 5401321917 | ОГРН 1095401001519
        </p>
        <p style={{marginTop: '15px', fontSize: '0.9rem'}}>
          <TelegramLink href="https://t.me/h2o_nsk_bot" target="_blank" rel="noopener noreferrer">
            Telegram бот: @h2o_nsk_bot
          </TelegramLink>
        </p>
        <p style={{marginTop: '15px', fontSize: '0.85rem'}}>
          <TelegramLink href="/oferta.html" target="_blank" rel="noopener noreferrer">
            Публичная оферта
          </TelegramLink>
        </p>
      </Footer>
    </HomeContainer>
  );
};

export default HomePage;
