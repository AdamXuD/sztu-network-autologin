import os
import time
import asyncio
from aiohttp import request, FormData, ClientTimeout
from dotenv import load_dotenv
from bs4 import BeautifulSoup

GETIP_URL = "http://www.qq.com/"
TEST_URL = "https://www.baidu.com/"
LOGIN_URL = (
    "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/cmcc_login/"
)
ONSUCCESS_URL = (
    "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/success/"
)
ONFAIL_URL = "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/fail/"
CHECKSTATUS_URL = (
    "http://47.98.217.39/lfradius/libs/portal/unify/portal.php/login/cmcc_login_result/"
)


async def isOnline():
    try:
        async with request("GET", TEST_URL, timeout=ClientTimeout(total=5)) as req:
            return req.status == 200
    except:
        return False


async def getNowIP():
    async with request("GET", GETIP_URL) as req:
        if len(req.history) == 0:
            return None
        return req.url.query["wlanuserip"]


def parseTokenFromHtml(html: str):
    soup = BeautifulSoup(html, "html.parser")
    element = soup.select_one("form input")
    return element.get("value")


async def postLogin(usrip: str):
    data = FormData(
        {
            "usrname": os.getenv("USER_ID"),
            "passwd": os.getenv("PASSWORD"),
            "treaty": "on",
            "nasid": 3,
            "usrmac": os.getenv("DEVICE_MAC"),
            "usrip": usrip,
            "basip": "172.17.127.254",
            "success": ONSUCCESS_URL,
            "fail": ONFAIL_URL,
            "offline": 0,
            "portal_version": 1,
            "portal_papchap": "pap",
        }
    )
    async with request(
        "POST",
        LOGIN_URL,
        data=data,
        headers={"Cookie": f"portal_usrname={os.getenv('USER_ID')}"},
    ) as req:
        if req.status == 200:
            text = await req.text()
            return parseTokenFromHtml(text)
        else:
            return None


async def postToken(token: str):
    async with request(
        "POST",
        LOGIN_URL,
        data=FormData({"cmcc_login_value": token}),
        headers={"Cookie": f"portal_usrname={os.getenv('USER_ID')}"},
    ) as req:
        if req.status == 200:
            await req.text()
            return True
        else:
            return False


async def getStatus(token: str):
    count = 0
    while count < 5:
        async with request(
            "POST",
            CHECKSTATUS_URL,
            data=FormData({"l": token}),
            headers={"Cookie": f"portal_usrname={os.getenv('USER_ID')}"},
        ) as req:
            text = await req.text()
            if text == "success":
                return True
        time.sleep(0.5)
    return False


async def checkStatus():
    async with request("GET", ONSUCCESS_URL) as req:
        if req.status == 200:
            return True
        else:
            return False


async def login():
    usrip = await getNowIP()
    if not usrip:
        return False, "无法获取用户IP。"

    token = await postLogin(usrip)
    if not token:
        return False, "无法正确获取Token。"

    if not await postToken(token):
        return False, "无法提交登录状态。"

    if not await getStatus(token):
        return False, "登录状态失败。"

    if not await checkStatus():
        return False, "登录失败。"
    return True, "登录成功。"


def printLog(hint: str):
    timeStr = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(time.time()))
    print(f"{timeStr} | {hint}")


def main():
    count = 0
    checkInterval = int(os.getenv("CHECK_INTERVAL"))
    retryMaxCount = int(os.getenv("RETRY_MAXCOUNT"))
    eventLoop = asyncio.get_event_loop()
    printLog("网络离线检测已启动。")
    while count < retryMaxCount:
        while eventLoop.run_until_complete(isOnline()):
            time.sleep(checkInterval)
        printLog(f"网络离线，正在进行第{count + 1}次重连... ...")
        res, hint = eventLoop.run_until_complete(login())
        count = 0 if res else count + 1
        printLog(hint)
        time.sleep(600)


if __name__ == "__main__":
    load_dotenv()
    try:
        main()
    except:
        pass
