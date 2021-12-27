import requests
from lxml import etree
import wget

def main():
    index = 684
    try:
        for i in range(684,2007):
            r = requests.get("https://service.mercurycom.com.cn/download-"+str(i)+".html")
            html = etree.HTML(r.content)
            result = html.xpath("//tr/td[2]/a/@href")
            url = ','.join(result)
            file = wget.download("https://service.mercurycom.com.cn"+url)
            print(file)
            # print("https://service.mercurycom.com.cn"+url)
            index += 1
            print(index)
    except Exception:
        pass

if __name__ =='__main__':
    main()