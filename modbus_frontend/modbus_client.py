#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Modbus TCP клиент для работы с ОВЕН ПЛК210
Автор: Alexandr
"""

import socket
import struct
import time
import logging
from typing import List, Tuple, Optional, Union
from dataclasses import dataclass

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('modbus_client.log', encoding='utf-8'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class ModbusConfig:
    """Конфигурация Modbus TCP соединения"""
    host: str = "192.168.1.100"  # IP адрес ПЛК210
    port: int = 502              # Стандартный порт Modbus TCP
    unit_id: int = 1             # Unit ID устройства
    timeout: float = 5.0         # Таймаут соединения в секундах
    retry_count: int = 3         # Количество попыток переподключения

class ModbusTCPClient:
    """Modbus TCP клиент для работы с ОВЕН ПЛК210"""
    
    def __init__(self, config: ModbusConfig):
        self.config = config
        self.socket = None
        self.transaction_id = 0
        
    def connect(self) -> bool:
        """Установка соединения с ПЛК"""
        try:
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.settimeout(self.config.timeout)
            self.socket.connect((self.config.host, self.config.port))
            logger.info(f"Соединение с ПЛК210 установлено: {self.config.host}:{self.config.port}")
            return True
        except Exception as e:
            logger.error(f"Ошибка подключения к ПЛК210: {e}")
            return False
    
    def disconnect(self):
        """Закрытие соединения"""
        if self.socket:
            self.socket.close()
            self.socket = None
            logger.info("Соединение с ПЛК210 закрыто")
    
    def _get_transaction_id(self) -> int:
        """Получение следующего ID транзакции"""
        self.transaction_id = (self.transaction_id + 1) % 65536
        return self.transaction_id
    
    def _send_request(self, function_code: int, data: bytes) -> Optional[bytes]:
        """Отправка запроса и получение ответа"""
        if not self.socket:
            logger.error("Соединение не установлено")
            return None
        
        # Формирование Modbus TCP заголовка
        transaction_id = self._get_transaction_id()
        protocol_id = 0  # Modbus
        length = len(data) + 2  # +2 для unit_id и function_code
        
        header = struct.pack('>HHH', transaction_id, protocol_id, length)
        request = header + struct.pack('B', self.config.unit_id) + struct.pack('B', function_code) + data
        
        try:
            # Отправка запроса
            self.socket.send(request)
            logger.debug(f"Отправлен запрос: {request.hex()}")
            
            # Получение ответа
            response = self.socket.recv(1024)
            logger.debug(f"Получен ответ: {response.hex()}")
            
            # Проверка заголовка ответа
            if len(response) < 8:
                logger.error("Слишком короткий ответ")
                return None
            
            resp_transaction_id, resp_protocol_id, resp_length = struct.unpack('>HHH', response[:6])
            
            if resp_transaction_id != transaction_id:
                logger.error(f"Неверный ID транзакции: ожидался {transaction_id}, получен {resp_transaction_id}")
                return None
            
            if resp_protocol_id != 0:
                logger.error(f"Неверный ID протокола: {resp_protocol_id}")
                return None
            
            # Проверка на ошибку
            if response[7] & 0x80:  # Флаг ошибки
                error_code = response[8]
                logger.error(f"Modbus ошибка: код {error_code}")
                return None
            
            return response[6:]  # Возвращаем данные без заголовка
            
        except socket.timeout:
            logger.error("Таймаут при получении ответа")
            return None
        except Exception as e:
            logger.error(f"Ошибка при отправке/получении данных: {e}")
            return None
    
    def read_holding_registers(self, start_address: int, count: int) -> Optional[List[int]]:
        """
        Чтение holding registers (функция 03)
        
        Args:
            start_address: Начальный адрес регистра
            count: Количество регистров для чтения
            
        Returns:
            Список значений регистров или None при ошибке
        """
        data = struct.pack('>HH', start_address, count)
        response = self._send_request(3, data)
        
        if response is None:
            return None
        
        if len(response) < 3:
            logger.error("Неверный формат ответа")
            return None
        
        byte_count = response[1]
        if len(response) < 2 + byte_count:
            logger.error("Неполный ответ")
            return None
        
        values = []
        for i in range(0, byte_count, 2):
            if i + 2 <= byte_count:
                value = struct.unpack('>H', response[2+i:4+i])[0]
                values.append(value)
        
        logger.info(f"Прочитано {len(values)} holding registers начиная с адреса {start_address}")
        return values
    
    def read_input_registers(self, start_address: int, count: int) -> Optional[List[int]]:
        """
        Чтение input registers (функция 04)
        
        Args:
            start_address: Начальный адрес регистра
            count: Количество регистров для чтения
            
        Returns:
            Список значений регистров или None при ошибке
        """
        data = struct.pack('>HH', start_address, count)
        response = self._send_request(4, data)
        
        if response is None:
            return None
        
        if len(response) < 3:
            logger.error("Неверный формат ответа")
            return None
        
        byte_count = response[1]
        if len(response) < 2 + byte_count:
            logger.error("Неполный ответ")
            return None
        
        values = []
        for i in range(0, byte_count, 2):
            if i + 2 <= byte_count:
                value = struct.unpack('>H', response[2+i:4+i])[0]
                values.append(value)
        
        logger.info(f"Прочитано {len(values)} input registers начиная с адреса {start_address}")
        return values
    
    def read_coils(self, start_address: int, count: int) -> Optional[List[bool]]:
        """
        Чтение coils (функция 01)
        
        Args:
            start_address: Начальный адрес coil
            count: Количество coils для чтения
            
        Returns:
            Список значений coils или None при ошибке
        """
        data = struct.pack('>HH', start_address, count)
        response = self._send_request(1, data)
        
        if response is None:
            return None
        
        if len(response) < 3:
            logger.error("Неверный формат ответа")
            return None
        
        byte_count = response[1]
        if len(response) < 2 + byte_count:
            logger.error("Неполный ответ")
            return None
        
        values = []
        for byte in response[2:2+byte_count]:
            for bit in range(8):
                if len(values) < count:
                    # Правильная интерпретация битов Modbus
                    # В Modbus биты упакованы в байты, где первый coil - младший бит
                    values.append(bool(byte & (1 << bit)))
        
        logger.info(f"Прочитано {len(values)} coils начиная с адреса {start_address}")
        return values
    
    def read_discrete_inputs(self, start_address: int, count: int) -> Optional[List[bool]]:
        """
        Чтение discrete inputs (функция 02)
        
        Args:
            start_address: Начальный адрес discrete input
            count: Количество discrete inputs для чтения
            
        Returns:
            Список значений discrete inputs или None при ошибке
        """
        data = struct.pack('>HH', start_address, count)
        response = self._send_request(2, data)
        
        if response is None:
            return None
        
        if len(response) < 3:
            logger.error("Неверный формат ответа")
            return None
        
        byte_count = response[1]
        if len(response) < 2 + byte_count:
            logger.error("Неполный ответ")
            return None
        
        values = []
        for byte in response[2:2+byte_count]:
            for bit in range(8):
                if len(values) < count:
                    # Правильная интерпретация битов Modbus
                    # В Modbus биты упакованы в байты, где первый discrete input - младший бит
                    values.append(bool(byte & (1 << bit)))
        
        logger.info(f"Прочитано {len(values)} discrete inputs начиная с адреса {start_address}")
        return values
    
    def write_single_coil(self, address: int, value: bool) -> bool:
        """
        Запись одного coil (функция 05)
        
        Args:
            address: Адрес coil
            value: Значение для записи (True/False)
            
        Returns:
            True при успехе, False при ошибке
        """
        # В Modbus для записи coil используется 0xFF00 для True и 0x0000 для False
        coil_value = 0xFF00 if value else 0x0000
        data = struct.pack('>HH', address, coil_value)
        
        # Создаем полный запрос
        transaction_id = 1
        request = struct.pack('>HHHBB', transaction_id, 0, 6, self.config.unit_id, 5) + data
        
        try:
            # Отправка запроса
            self.socket.send(request)
            logger.debug(f"Отправлен запрос записи coil: {request.hex()}")
            
            # Получение ответа
            response = self.socket.recv(1024)
            logger.debug(f"Получен ответ записи coil: {response.hex()}")
            
            # Проверка заголовка ответа
            if len(response) < 12:
                logger.error("Слишком короткий ответ при записи coil")
                return False
            
            resp_transaction_id, resp_protocol_id, resp_length = struct.unpack('>HHH', response[:6])
            
            if resp_transaction_id != transaction_id:
                logger.error(f"Неверный ID транзакции при записи coil: ожидался {transaction_id}, получен {resp_transaction_id}")
                return False
            
            if resp_protocol_id != 0:
                logger.error(f"Неверный ID протокола при записи coil: {resp_protocol_id}")
                return False
            
            # Проверка на ошибку
            if response[7] & 0x80:  # Флаг ошибки
                error_code = response[8]
                logger.error(f"Modbus ошибка при записи coil: код {error_code}")
                return False
            
            # Проверяем, что ответ содержит те же данные, что мы отправили
            resp_address, resp_value = struct.unpack('>HH', response[8:12])
            if resp_address == address and resp_value == coil_value:
                logger.info(f"Записан coil {address} = {value} ({'ВКЛ' if value else 'ВЫКЛ'})")
                return True
            else:
                logger.error(f"Ошибка записи coil {address}: ожидался адрес {address}, значение {coil_value}, получен адрес {resp_address}, значение {resp_value}")
                return False
                
        except socket.timeout:
            logger.error("Таймаут при записи coil")
            return False
        except Exception as e:
            logger.error(f"Ошибка при записи coil: {e}")
            return False
    
    def write_multiple_coils(self, start_address: int, values: List[bool]) -> bool:
        """
        Запись нескольких coils (функция 15)
        
        Args:
            start_address: Начальный адрес coil
            values: Список значений для записи (True/False)
            
        Returns:
            True при успехе, False при ошибке
        """
        # Вычисляем количество байтов для упаковки coils
        byte_count = (len(values) + 7) // 8
        
        # Упаковываем coils в байты
        packed_data = bytearray(byte_count)
        for i, value in enumerate(values):
            if value:
                byte_index = i // 8
                bit_index = i % 8
                packed_data[byte_index] |= (1 << bit_index)
        
        data = struct.pack('>HHB', start_address, len(values), byte_count)
        data += bytes(packed_data)
        
        response = self._send_request(15, data)
        
        if response is None:
            return False
        
        if len(response) < 5:
            logger.error("Неверный формат ответа")
            return False
        
        # В Modbus TCP ответ: [transaction_id(2), protocol_id(2), length(2), unit_id(1), function_code(1), address(2), count(2)]
        if len(response) >= 12:
            resp_address, resp_count = struct.unpack('>HH', response[8:12])
            if resp_address == start_address and resp_count == len(values):
                logger.info(f"Записано {len(values)} coils начиная с адреса {start_address}")
                return True
            else:
                logger.error(f"Ошибка записи coils: ожидался адрес {start_address}, количество {len(values)}, получен адрес {resp_address}, количество {resp_count}")
                return False
        else:
            logger.error("Неполный ответ при записи coils")
            return False
    
    def write_single_register(self, address: int, value: int) -> bool:
        """
        Запись одного holding register (функция 06)
        
        Args:
            address: Адрес регистра
            value: Значение для записи
            
        Returns:
            True при успехе, False при ошибке
        """
        data = struct.pack('>HH', address, value)
        response = self._send_request(6, data)
        
        if response is None:
            return False
        
        if len(response) < 5:
            logger.error("Неверный формат ответа")
            return False
        
        # Проверяем, что ответ содержит те же данные, что мы отправили
        # В Modbus TCP ответ: [transaction_id(2), protocol_id(2), length(2), unit_id(1), function_code(1), address(2), value(2)]
        if len(response) >= 12:
            resp_address, resp_value = struct.unpack('>HH', response[8:12])
            if resp_address == address and resp_value == value:
                logger.info(f"Записан holding register {address} = {value}")
                return True
            else:
                logger.error(f"Ошибка записи holding register: ожидался адрес {address}, значение {value}, получен адрес {resp_address}, значение {resp_value}")
                return False
        else:
            logger.error("Неполный ответ при записи holding register")
            return False
    
    def write_multiple_registers(self, start_address: int, values: List[int]) -> bool:
        """
        Запись нескольких holding registers (функция 16)
        
        Args:
            start_address: Начальный адрес регистра
            values: Список значений для записи
            
        Returns:
            True при успехе, False при ошибке
        """
        byte_count = len(values) * 2
        data = struct.pack('>HHB', start_address, len(values), byte_count)
        
        for value in values:
            data += struct.pack('>H', value)
        
        response = self._send_request(16, data)
        
        if response is None:
            return False
        
        if len(response) < 5:
            logger.error("Неверный формат ответа")
            return False
        
        # В Modbus TCP ответ: [transaction_id(2), protocol_id(2), length(2), unit_id(1), function_code(1), address(2), count(2)]
        if len(response) >= 12:
            resp_address, resp_count = struct.unpack('>HH', response[8:12])
            if resp_address == start_address and resp_count == len(values):
                logger.info(f"Записано {len(values)} holding registers начиная с адреса {start_address}")
                return True
            else:
                logger.error(f"Ошибка записи holding registers: ожидался адрес {start_address}, количество {len(values)}, получен адрес {resp_address}, количество {resp_count}")
                return False
        else:
            logger.error("Неполный ответ при записи holding registers")
            return False

def main():
    """Пример использования Modbus TCP клиента"""
    
    # Конфигурация для ПЛК210
    config = ModbusConfig(
        host="192.168.1.100",  # Замените на IP вашего ПЛК210
        port=502,
        unit_id=1,
        timeout=5.0
    )
    
    client = ModbusTCPClient(config)
    
    try:
        # Подключение к ПЛК
        if not client.connect():
            logger.error("Не удалось подключиться к ПЛК210")
            return
        
        # Примеры чтения данных
        logger.info("=== Примеры работы с ПЛК210 ===")
        
        # Чтение holding registers (обычно используются для конфигурации)
        holding_regs = client.read_holding_registers(0, 10)
        if holding_regs:
            logger.info(f"Holding registers 0-9: {holding_regs}")
        
        # Чтение input registers (обычно используются для входных данных)
        input_regs = client.read_input_registers(0, 10)
        if input_regs:
            logger.info(f"Input registers 0-9: {input_regs}")
        
        # Чтение coils (дискретные выходы)
        coils = client.read_coils(0, 16)
        if coils:
            logger.info(f"Coils 0-15: {coils}")
        
        # Пример записи данных
        success = client.write_single_register(100, 1234)
        if success:
            logger.info("Запись в регистр 100 выполнена успешно")
        
        # Чтение записанного значения
        written_value = client.read_holding_registers(100, 1)
        if written_value:
            logger.info(f"Прочитано значение из регистра 100: {written_value[0]}")
        
    except KeyboardInterrupt:
        logger.info("Работа прервана пользователем")
    except Exception as e:
        logger.error(f"Неожиданная ошибка: {e}")
    finally:
        client.disconnect()

if __name__ == "__main__":
    main()
