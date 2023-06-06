#!/usr/bin/env python3

import re
import sys
import datetime
from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager


email = sys.argv[1]
password = sys.argv[2]
bank_name = sys.argv[3]

options = Options()
options.add_argument("--headless")
options.add_argument("--disable-dev-shm-usage")
driver = webdriver.Chrome(
    service=Service(ChromeDriverManager().install()), options=options
)

driver.get("https://web.budgetbakers.com/")

email_field = WebDriverWait(driver, 20).until(
    EC.presence_of_element_located((By.NAME, "email"))
)
email_field.send_keys(email)

password_field = WebDriverWait(driver, 20).until(
    EC.presence_of_element_located((By.NAME, "password"))
)
password_field.send_keys(password)

login_button = driver.find_element(By.CSS_SELECTOR, '.ui.large.circular.fluid.primary.button')
login_button.click()

records_link = WebDriverWait(driver, 20).until(
    EC.visibility_of_element_located((By.LINK_TEXT, "Records"))
)
records_link.click()

last_30_days_button = WebDriverWait(driver, 20).until(
    EC.element_to_be_clickable((By.XPATH, "//div[@class='selection-inner' and text()='Last 30 days']"))
)
last_30_days_button.click()

all_checkbox = WebDriverWait(driver, 20).until(
    EC.visibility_of_element_located((By.XPATH, "//div[@class='ui radio checkbox']/label[text()='All']"))
)
all_checkbox.click()

driver.find_element(By.XPATH, f"//span[contains(text(),'{bank_name}')]").click()

div_elements = WebDriverWait(driver, 20).until(EC.presence_of_all_elements_located((By.TAG_NAME, "div")))

last_date = None
for div_element in div_elements:
    if re.search(r"^[A-Za-z]+\s\d{1,2}$", div_element.text):
        last_date = div_element.text
        break


print(last_date)    

driver.quit()
