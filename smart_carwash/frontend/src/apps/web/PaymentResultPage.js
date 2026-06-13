import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import styled from 'styled-components';

const Page = styled.div`
  padding: 40px 24px;
  text-align: center;
  max-width: 480px;
  margin: 0 auto;
  @media (min-width: 600px) {
    padding: 56px 40px;
    max-width: 520px;
  }
`;

const Title = styled.h2`
  color: ${(p) => (p.success ? '#2e7d32' : '#c62828')};
  margin-bottom: 16px;
  font-size: 1.35rem;
  @media (min-width: 600px) {
    font-size: 1.5rem;
  }
`;

const WebPaymentResultPage = ({ success }) => {
  const navigate = useNavigate();
  useEffect(() => {
    const t = setTimeout(() => navigate('/web'), 5000);
    return () => clearTimeout(t);
  }, [navigate]);

  return (
    <Page>
      <Title success={success}>
        {success ? 'Оплата прошла успешно' : 'Оплата не выполнена'}
      </Title>
      <p>{success ? 'Спасибо! Вы будете перенаправлены на главную.' : 'Попробуйте снова или обратитесь в поддержку.'}</p>
      <p style={{ marginTop: 24 }}>
        <button onClick={() => navigate('/web')} style={{ padding: '10px 20px', fontSize: 16, cursor: 'pointer' }}>
          На главную
        </button>
      </p>
    </Page>
  );
};

export default WebPaymentResultPage;
