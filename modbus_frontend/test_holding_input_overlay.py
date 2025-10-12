#!/usr/bin/env python3
"""
Тест overlay Holding и Input Registers
По запросу Кирилла: записать в Holding 0,1 и прочитать Input 0,1
"""

import socket
import struct
import time
import logging
from typing import List, Optional
from dataclasses import dataclass

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

@dataclass
class ModbusConfig:
    host: str = "195.208.131.189"
    port: int = 502
    unit_id: int = 1
    timeout: float = 5.0

class ModbusTCPClient:
    def __init__(self, config: ModbusConfig):
        self.config = config
        self.socket = None
        self.transaction_id = 1
    
    def connect(self) -> bool:
        try:
            if self.socket:
                self.socket.close()
            
            self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.socket.settimeout(self.config.timeout)
            self.socket.connect((self.config.host, self.config.port))
            logger.info(f"✓ Подключение к ПЛК210: {self.config.host}:{self.config.port}")
            return True
        except Exception as e:
            logger.error(f"✗ Ошибка подключения: {e}")
            return False
    
    def disconnect(self):
        if self.socket:
            try:
                self.socket.close()
            except:
                pass
            finally:
                self.socket = None
                logger.info("Отключение от ПЛК210")
    
    def _send_request(self, data: bytes) -> Optional[bytes]:
        try:
            self.socket.send(data)
            response = self.socket.recv(1024)
            return response
        except Exception as e:
            logger.error(f"Ошибка отправки запроса: {e}")
            return None
    
    def _create_modbus_header(self, pdu_length: int) -> bytes:
        self.transaction_id = (self.transaction_id % 65535) + 1
        return struct.pack('>HHHB', 
                          self.transaction_id,
                          0,
                          pdu_length + 1,
                          self.config.unit_id)
    
    def read_holding_registers(self, address: int, count: int) -> Optional[List[int]]:
        """Чтение Holding Registers (Function 0x03)"""
        try:
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x03, address, count)
            request = header + pdu
            
            response = self._send_request(request)
            if not response or len(response) < 9:
                logger.error(f"Неполный ответ при чтении Holding Registers {address}")
                return None
            
            byte_count = response[8]
            if len(response) < 9 + byte_count:
                logger.error(f"Недостаточно данных в ответе")
                return None
            
            values = []
            for i in range(count):
                offset = 9 + i * 2
                value = struct.unpack('>H', response[offset:offset+2])[0]
                values.append(value)
            
            return values
        except Exception as e:
            logger.error(f"Ошибка чтения Holding Registers {address}: {e}")
            return None
    
    def read_input_registers(self, address: int, count: int) -> Optional[List[int]]:
        """Чтение Input Registers (Function 0x04)"""
        try:
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x04, address, count)
            request = header + pdu
            
            response = self._send_request(request)
            if not response or len(response) < 9:
                logger.error(f"Неполный ответ при чтении Input Registers {address}")
                return None
            
            byte_count = response[8]
            if len(response) < 9 + byte_count:
                logger.error(f"Недостаточно данных в ответе")
                return None
            
            values = []
            for i in range(count):
                offset = 9 + i * 2
                value = struct.unpack('>H', response[offset:offset+2])[0]
                values.append(value)
            
            return values
        except Exception as e:
            logger.error(f"Ошибка чтения Input Registers {address}: {e}")
            return None
    
    def write_single_register(self, address: int, value: int) -> bool:
        """Запись в один Holding Register (Function 0x06)"""
        try:
            header = self._create_modbus_header(5)
            pdu = struct.pack('>BHH', 0x06, address, value)
            request = header + pdu
            
            response = self._send_request(request)
            if not response:
                logger.error(f"Нет ответа при записи в Register {address}")
                return False
            
            # Для записи регистра ПЛК может отвечать не полностью
            logger.info(f"✓ Запись в Holding Register {address} = {value}")
            return True
        except Exception as e:
            logger.error(f"Ошибка записи в Holding Register {address}: {e}")
            return False

def test_overlay_registers():
    """Тест overlay Holding и Input Registers по запросу Кирилла"""
    print("🔄 ТЕСТ OVERLAY HOLDING + INPUT REGISTERS")
    print("=" * 60)
    
    config = ModbusConfig()
    client = ModbusTCPClient(config)
    
    if not client.connect():
        print("❌ Не удалось подключиться к ПЛК210")
        return
    
    try:
        # Тестовые значения
        test_values = [0x1234, 0x5678]  # Для Register 0 и 1
        
        print(f"\n📝 ПЛАН ТЕСТА:")
        print(f"1. Записать в Holding Register 0 = {test_values[0]} (0x{test_values[0]:04X})")
        print(f"2. Записать в Holding Register 1 = {test_values[1]} (0x{test_values[1]:04X})")
        print(f"3. Прочитать Holding Registers 0-1")
        print(f"4. Прочитать Input Registers 0-1")
        print(f"5. Сравнить результаты")
        
        # Шаг 1: Запись в Holding Register 0
        print(f"\n🔧 ШАГ 1: Запись в Holding Register 0")
        success = client.write_single_register(0, test_values[0])
        print(f"Результат записи Register 0: {'✓ Успешно' if success else '✗ Ошибка'}")
        
        time.sleep(0.5)
        
        # Шаг 2: Запись в Holding Register 1
        print(f"\n🔧 ШАГ 2: Запись в Holding Register 1")
        success = client.write_single_register(1, test_values[1])
        print(f"Результат записи Register 1: {'✓ Успешно' if success else '✗ Ошибка'}")
        
        time.sleep(1)
        
        # Шаг 3: Чтение Holding Registers 0-1
        print(f"\n📖 ШАГ 3: Чтение Holding Registers 0-1")
        holding_values = client.read_holding_registers(0, 2)
        if holding_values:
            print(f"✓ Holding Register 0: {holding_values[0]} (0x{holding_values[0]:04X})")
            print(f"✓ Holding Register 1: {holding_values[1]} (0x{holding_values[1]:04X})")
        else:
            print("✗ Ошибка чтения Holding Registers")
        
        # Шаг 4: Чтение Input Registers 0-1
        print(f"\n📖 ШАГ 4: Чтение Input Registers 0-1")
        input_values = client.read_input_registers(0, 2)
        if input_values:
            print(f"✓ Input Register 0: {input_values[0]} (0x{input_values[0]:04X})")
            print(f"✓ Input Register 1: {input_values[1]} (0x{input_values[1]:04X})")
        else:
            print("✗ Ошибка чтения Input Registers")
        
        # Шаг 5: Анализ результатов
        print(f"\n🔍 ШАГ 5: АНАЛИЗ РЕЗУЛЬТАТОВ")
        print("=" * 60)
        
        if holding_values and input_values:
            print("📊 СРАВНЕНИЕ ЗНАЧЕНИЙ:")
            
            for i in range(2):
                holding_val = holding_values[i]
                input_val = input_values[i]
                written_val = test_values[i]
                
                print(f"\nRegister {i}:")
                print(f"  Записано в Holding: {written_val} (0x{written_val:04X})")
                print(f"  Прочитано Holding:  {holding_val} (0x{holding_val:04X}) {'✓' if holding_val == written_val else '✗'}")
                print(f"  Прочитано Input:    {input_val} (0x{input_val:04X}) {'✓' if input_val == written_val else '✗'}")
                
                if holding_val == input_val == written_val:
                    print(f"  🎯 OVERLAY РАБОТАЕТ! Все значения совпадают")
                elif holding_val == input_val:
                    print(f"  ⚠️  Holding = Input, но не равны записанному значению")
                elif holding_val == written_val:
                    print(f"  ⚠️  Holding корректен, но Input отличается")
                elif input_val == written_val:
                    print(f"  ⚠️  Input корректен, но Holding отличается")
                else:
                    print(f"  ❌ Все значения разные")
            
            # Общий вывод
            print(f"\n🎯 ОБЩИЙ РЕЗУЛЬТАТ:")
            overlay_works = all(
                holding_values[i] == input_values[i] 
                for i in range(2)
            )
            
            if overlay_works:
                print("✅ OVERLAY ФУНКЦИОНИРУЕТ!")
                print("   Holding и Input Registers возвращают одинаковые значения")
                
                values_match = all(
                    holding_values[i] == test_values[i] 
                    for i in range(2)
                )
                
                if values_match:
                    print("✅ ЗАПИСЬ РАБОТАЕТ КОРРЕКТНО!")
                    print("   Записанные значения сохраняются правильно")
                else:
                    print("⚠️  Запись работает частично")
                    print("   Значения сохраняются, но могут изменяться")
            else:
                print("❌ OVERLAY НЕ РАБОТАЕТ")
                print("   Holding и Input Registers возвращают разные значения")
        
        elif holding_values:
            print("⚠️  Holding Registers читаются, но Input Registers недоступны")
            print("   Возможно, overlay не настроен в CODESYS")
            
        elif input_values:
            print("⚠️  Input Registers читаются, но Holding Registers недоступны")
            print("   Это странная ситуация, требует проверки")
            
        else:
            print("❌ Ни Holding, ни Input Registers не читаются")
            print("   Проблема с конфигурацией Modbus в CODESYS")
    
    finally:
        client.disconnect()
    
    print(f"\n🏁 ТЕСТ ЗАВЕРШЕН")
    print("=" * 60)

def main():
    print("🏭 ПЛК210 - Тест Overlay Holding + Input Registers")
    print("Запрос от Кирилла: записать в Holding 0,1 и прочитать Input 0,1")
    print()
    
    try:
        test_overlay_registers()
    except KeyboardInterrupt:
        print("\n⏹️  Тест прерван пользователем")
    except Exception as e:
        print(f"\n❌ Ошибка теста: {e}")

if __name__ == "__main__":
    main()
