import React from 'react';
import styled from 'styled-components';

const WashInfoContainer = styled.div`
  margin-bottom: 24px;
`;

const Title = styled.h2`
  font-size: 20px;
  margin-bottom: 16px;
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
`;

const BoxesGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 16px;
`;

const BoxCard = styled.div`
  padding: 16px;
  border-radius: 12px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease, background-color 0.3s ease;
  
  &:hover {
    transform: translateY(-4px);
  }
`;

const BoxNumber = styled.h3`
  font-size: 18px;
  margin-bottom: 8px;
  color: ${props => props.theme === 'dark' ? '#FFFFFF' : '#000000'};
`;

const BoxStatus = styled.div`
  display: inline-block;
  padding: 6px 12px;
  border-radius: 16px;
  font-size: 14px;
  font-weight: 500;
  background-color: ${props => {
    if (props.status === 'free') return props.theme === 'dark' ? '#2E7D32' : '#E8F5E9';
    if (props.status === 'busy') return props.theme === 'dark' ? '#C62828' : '#FFEBEE';
    return props.theme === 'dark' ? '#F57C00' : '#FFF3E0';
  }};
  color: ${props => {
    if (props.status === 'free') return props.theme === 'dark' ? '#FFFFFF' : '#2E7D32';
    if (props.status === 'busy') return props.theme === 'dark' ? '#FFFFFF' : '#C62828';
    return props.theme === 'dark' ? '#FFFFFF' : '#F57C00';
  }};
`;

const NoBoxesMessage = styled.p`
  font-size: 16px;
  color: ${props => props.theme === 'dark' ? '#E0E0E0' : '#666666'};
  text-align: center;
  padding: 24px;
  background-color: ${props => props.theme === 'dark' ? '#2C2C2C' : '#FFFFFF'};
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
`;

const getStatusText = (status) => {
  switch (status) {
    case 'free':
      return 'Свободен';
    case 'busy':
      return 'Занят';
    case 'maintenance':
      return 'На обслуживании';
    default:
      return 'Неизвестно';
  }
};

const WashInfo = ({ washInfo, theme }) => {
  const boxes = washInfo?.boxes || [];

  return (
    <WashInfoContainer>
      <Title theme={theme}>Боксы автомойки</Title>
      
      {boxes.length > 0 ? (
        <BoxesGrid>
          {boxes.map((box) => (
            <BoxCard key={box.id} theme={theme}>
              <BoxNumber theme={theme}>Бокс #{box.number}</BoxNumber>
              <BoxStatus status={box.status} theme={theme}>
                {getStatusText(box.status)}
              </BoxStatus>
            </BoxCard>
          ))}
        </BoxesGrid>
      ) : (
        <NoBoxesMessage theme={theme}>
          Информация о боксах отсутствует
        </NoBoxesMessage>
      )}
    </WashInfoContainer>
  );
};

export default WashInfo;
