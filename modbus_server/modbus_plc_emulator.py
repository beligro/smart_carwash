#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Эмулятор ПЛК для тестирования Modbus TCP
Работает как Modbus TCP сервер на порту 502

Автор: Smart Car Wash System
"""

import socket
import struct
import threading
import time
import logging
from typing import Dict, List
from dataclasses import dataclass

# Настройка логирования
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('modbus_plc_emulator.log', encoding='utf-8'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class ModbusServerConfig:
    """Конфигурация Modbus TCP сервера"""
    host: str = "0.0.0.0"  # Слушаем на всех интерфейсах
    port: int = 5502       # Альтернативный порт Modbus TCP (502 может требовать root)
    unit_id: int = 1       # Unit ID устройства

class ModbusPLCEmulator:
    """Эмулятор ПЛК с Modbus TCP сервером"""
    
    def __init__(self, config: ModbusServerConfig):
        self.config = config
        self.running = False
        self.server_socket = None
        self.clients = []
        
        # Эмуляция регистров ПЛК (6 регистров по 16 бит)
        self.holding_registers = [0] * 6  # Регистры 0-5
        self.coils = [False] * 96         # 96 coils (6 регистров * 16 бит)
        
        # Блокировка для thread-safe операций
        self.lock = threading.Lock()
        
        logger.info(f"Эмулятор ПЛК инициализирован: {config.host}:{config.port}")
    
    def start(self):
        """Запуск Modbus TCP сервера"""
        try:
            self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind((self.config.host, self.config.port))
            self.server_socket.listen(5)
            
            self.running = True
            logger.info(f"🚀 Modbus TCP сервер запущен на {self.config.host}:{self.config.port}")
            
            while self.running:
                try:
                    client_socket, client_address = self.server_socket.accept()
                    logger.info(f"Новое подключение от {client_address}")
                    
                    # Запускаем обработку клиента в отдельном потоке
                    client_thread = threading.Thread(
                        target=self.handle_client,
                        args=(client_socket, client_address)
                    )
                    client_thread.daemon = True
                    client_thread.start()
                    
                except socket.error as e:
                    if self.running:
                        logger.error(f"Ошибка сокета: {e}")
                    
        except Exception as e:
            logger.error(f"Ошибка запуска сервера: {e}")
        finally:
            self.stop()
    
    def stop(self):
        """Остановка сервера"""
        self.running = False
        if self.server_socket:
            self.server_socket.close()
        logger.info("Modbus TCP сервер остановлен")
    
    def handle_client(self, client_socket: socket.socket, client_address):
        """Обработка клиента"""
        try:
            while self.running:
                # Получаем запрос
                data = client_socket.recv(1024)
                if not data:
                    break
                
                logger.debug(f"Получен запрос от {client_address}: {data.hex()}")
                
                # Обрабатываем Modbus запрос
                response = self.process_modbus_request(data)
                if response:
                    client_socket.send(response)
                    logger.debug(f"Отправлен ответ: {response.hex()}")
                
        except Exception as e:
            logger.error(f"Ошибка обработки клиента {client_address}: {e}")
        finally:
            client_socket.close()
            logger.info(f"Клиент {client_address} отключен")
    
    def process_modbus_request(self, data: bytes) -> bytes:
        """Обработка Modbus запроса"""
        if len(data) < 8:
            logger.error("Слишком короткий запрос")
            return None
        
        try:
            # Парсим Modbus TCP заголовок
            transaction_id, protocol_id, length = struct.unpack('>HHH', data[:6])
            unit_id = data[6]
            function_code = data[7]
            
            logger.info(f"Modbus запрос: transaction_id={transaction_id}, function_code={function_code}, unit_id={unit_id}")
            
            # Проверяем Unit ID
            if unit_id != self.config.unit_id:
                logger.warning(f"Неверный Unit ID: {unit_id}, ожидался {self.config.unit_id}")
                return self.create_error_response(transaction_id, function_code, 0x0B)  # Gateway Target Device Failed to Respond
            
            # Обрабатываем функции Modbus
            if function_code == 0x05:  # Write Single Coil
                return self.handle_write_single_coil(data, transaction_id)
            elif function_code == 0x01:  # Read Coils
                return self.handle_read_coils(data, transaction_id)
            elif function_code == 0x03:  # Read Holding Registers
                return self.handle_read_holding_registers(data, transaction_id)
            elif function_code == 0x06:  # Write Single Register
                return self.handle_write_single_register(data, transaction_id)
            else:
                logger.warning(f"Неподдерживаемая функция: {function_code}")
                return self.create_error_response(transaction_id, function_code, 0x01)  # Illegal Function
                
        except Exception as e:
            logger.error(f"Ошибка обработки запроса: {e}")
            return None
    
    def handle_write_single_coil(self, data: bytes, transaction_id: int) -> bytes:
        """Обработка записи одного coil (функция 0x05)"""
        if len(data) < 12:
            return self.create_error_response(transaction_id, 0x05, 0x03)  # Illegal Data Value
        
        # Парсим адрес и значение
        address, value = struct.unpack('>HH', data[8:12])
        
        with self.lock:
            if address < len(self.coils):
                # Modbus: 0xFF00 = ON, 0x0000 = OFF
                self.coils[address] = (value == 0xFF00)
                logger.info(f"Coil {address} установлен в {self.coils[address]} (value=0x{value:04X})")
                
                # Обновляем соответствующий регистр
                register_num = address // 16
                bit_num = address % 16
                if register_num < len(self.holding_registers):
                    if self.coils[address]:
                        self.holding_registers[register_num] |= (1 << bit_num)
                    else:
                        self.holding_registers[register_num] &= ~(1 << bit_num)
                
                # Создаем успешный ответ
                response_data = struct.pack('>HH', address, value)
                return self.create_response(transaction_id, 0x05, response_data)
            else:
                logger.error(f"Неверный адрес coil: {address}")
                return self.create_error_response(transaction_id, 0x05, 0x02)  # Illegal Data Address
    
    def handle_read_coils(self, data: bytes, transaction_id: int) -> bytes:
        """Обработка чтения coils (функция 0x01)"""
        if len(data) < 12:
            return self.create_error_response(transaction_id, 0x01, 0x03)
        
        address, quantity = struct.unpack('>HH', data[8:12])
        
        with self.lock:
            if address + quantity <= len(self.coils):
                # Упаковываем coils в байты
                byte_count = (quantity + 7) // 8
                coil_bytes = []
                
                for byte_idx in range(byte_count):
                    byte_value = 0
                    for bit_idx in range(8):
                        coil_idx = address + byte_idx * 8 + bit_idx
                        if coil_idx < address + quantity and coil_idx < len(self.coils):
                            if self.coils[coil_idx]:
                                byte_value |= (1 << bit_idx)
                    coil_bytes.append(byte_value)
                
                response_data = struct.pack('B', byte_count) + bytes(coil_bytes)
                logger.info(f"Чтение coils {address}-{address+quantity-1}: {[self.coils[i] for i in range(address, min(address+quantity, len(self.coils)))]}")
                return self.create_response(transaction_id, 0x01, response_data)
            else:
                return self.create_error_response(transaction_id, 0x01, 0x02)
    
    def handle_read_holding_registers(self, data: bytes, transaction_id: int) -> bytes:
        """Обработка чтения holding registers (функция 0x03)"""
        if len(data) < 12:
            return self.create_error_response(transaction_id, 0x03, 0x03)
        
        address, quantity = struct.unpack('>HH', data[8:12])
        
        with self.lock:
            if address + quantity <= len(self.holding_registers):
                response_data = struct.pack('B', quantity * 2)  # Byte count
                for i in range(quantity):
                    reg_value = self.holding_registers[address + i] if address + i < len(self.holding_registers) else 0
                    response_data += struct.pack('>H', reg_value)
                
                logger.info(f"Чтение регистров {address}-{address+quantity-1}: {self.holding_registers[address:address+quantity]}")
                return self.create_response(transaction_id, 0x03, response_data)
            else:
                return self.create_error_response(transaction_id, 0x03, 0x02)
    
    def handle_write_single_register(self, data: bytes, transaction_id: int) -> bytes:
        """Обработка записи одного регистра (функция 0x06)"""
        if len(data) < 12:
            return self.create_error_response(transaction_id, 0x06, 0x03)
        
        address, value = struct.unpack('>HH', data[8:12])
        
        with self.lock:
            if address < len(self.holding_registers):
                self.holding_registers[address] = value
                logger.info(f"Регистр {address} установлен в {value} (0x{value:04X})")
                
                # Обновляем соответствующие coils
                for bit in range(16):
                    coil_idx = address * 16 + bit
                    if coil_idx < len(self.coils):
                        self.coils[coil_idx] = bool(value & (1 << bit))
                
                response_data = struct.pack('>HH', address, value)
                return self.create_response(transaction_id, 0x06, response_data)
            else:
                return self.create_error_response(transaction_id, 0x06, 0x02)
    
    def create_response(self, transaction_id: int, function_code: int, data: bytes) -> bytes:
        """Создание успешного ответа"""
        length = len(data) + 2  # +2 для unit_id и function_code
        header = struct.pack('>HHH', transaction_id, 0, length)
        response = header + struct.pack('BB', self.config.unit_id, function_code) + data
        return response
    
    def create_error_response(self, transaction_id: int, function_code: int, error_code: int) -> bytes:
        """Создание ответа с ошибкой"""
        header = struct.pack('>HHH', transaction_id, 0, 3)  # Length = 3 (unit_id + error_function + error_code)
        error_function = function_code | 0x80  # Устанавливаем бит ошибки
        response = header + struct.pack('BBB', self.config.unit_id, error_function, error_code)
        return response
    
    def get_status(self) -> Dict:
        """Получение статуса эмулятора"""
        with self.lock:
            return {
                "running": self.running,
                "host": self.config.host,
                "port": self.config.port,
                "unit_id": self.config.unit_id,
                "holding_registers": self.holding_registers.copy(),
                "coils_summary": {
                    "total": len(self.coils),
                    "active": sum(1 for coil in self.coils if coil),
                    "registers": [
                        {
                            "register": i,
                            "value": self.holding_registers[i],
                            "hex": f"0x{self.holding_registers[i]:04X}",
                            "coils": self.coils[i*16:(i+1)*16]
                        }
                        for i in range(len(self.holding_registers))
                    ]
                }
            }

def main():
    """Главная функция"""
    print("🏭 Эмулятор ПЛК для автомойки")
    print("=" * 50)
    
    # Создаем конфигурацию
    config = ModbusServerConfig(
        host="0.0.0.0",  # Слушаем на всех интерфейсах
        port=5502,       # Альтернативный Modbus TCP порт
        unit_id=1
    )
    
    # Создаем эмулятор
    emulator = ModbusPLCEmulator(config)
    
    try:
        # Запускаем сервер в отдельном потоке
        server_thread = threading.Thread(target=emulator.start)
        server_thread.daemon = True
        server_thread.start()
        
        print(f"✅ Эмулятор ПЛК запущен на порту {config.port}")
        print(f"📡 Готов принимать Modbus TCP команды")
        print(f"🔧 Unit ID: {config.unit_id}")
        print(f"📊 Регистров: {len(emulator.holding_registers)}")
        print(f"🔘 Coils: {len(emulator.coils)}")
        print("")
        print("Для остановки нажмите Ctrl+C")
        print("")
        
        # Периодически выводим статус
        while True:
            time.sleep(10)
            if emulator.running:
                status = emulator.get_status()
                active_coils = status["coils_summary"]["active"]
                print(f"[{time.strftime('%H:%M:%S')}] Активных coils: {active_coils}/{len(emulator.coils)}")
                
                # Показываем состояние первых 3 регистров
                for reg_info in status["coils_summary"]["registers"][:3]:
                    active_bits = sum(1 for bit in reg_info["coils"] if bit)
                    print(f"  Регистр {reg_info['register']}: {reg_info['hex']} ({active_bits}/16 бит)")
            else:
                break
                
    except KeyboardInterrupt:
        print("\n⏹️  Остановка эмулятора...")
        emulator.stop()
        print("👋 Эмулятор остановлен")
    except Exception as e:
        logger.error(f"Критическая ошибка: {e}")
        emulator.stop()

if __name__ == '__main__':
    main()
