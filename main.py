import asyncio
import argparse
import aiofiles
import aiohttp
import ujson
from datetime import datetime
import pathlib
import os

parser = argparse.ArgumentParser()
parser.add_argument('--link', '-L', help='Enter the full link to the page. Example: '
                                   '"https://telegra.ph/What-Was-TON-And-Why-It-Is-Over-05-12"', type=str,
                    required=True)
parser.add_argument('--folder', '-F', help='Specify the folder where to extract images', type=pathlib.Path,
                    default=pathlib.Path().absolute())
parser.add_argument('--explicit', '-E', help='Show all messages', choices=['yes', 'y', 'true', '1'], default=False)
parser.add_argument('--mode', '-M', help='Choose mode to download.',
                    choices=['ordered', 'fast'], default='fast')


async def download_image(_url, folder, image_id=None):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"https://telegra.ph/file/{_url}") as response:
            assert response.status == 200
            print(f"Downloading {_url}") if parser.parse_args().explicit else None
            if not pathlib.Path(folder).exists():
                try:
                    pathlib.Path(folder).mkdir(parents=True, exist_ok=True)
                except OSError:
                    print(f"Creation of the directory {folder} failed") if parser.parse_args().explicit else None
                else:
                    print(f"Successfully created the directory {folder}") if parser.parse_args().explicit else None

            if parser.parse_args().mode == 'fast':
                path = pathlib.Path().joinpath(folder, _url)
            else:
                path = pathlib.Path().joinpath(f"{folder}/{image_id}_{_url}")
            async with aiofiles.open(path, 'wb') as f:
                print(f"Writing {_url}") if parser.parse_args().explicit else None
                await f.write(await response.read())
                await f.flush()


async def main():
    async with aiohttp.ClientSession() as session:
        params = {'return_content': 'true'}
        async with session.get(
                f"https://api.telegra.ph/getPage/{parser.parse_args().link.lstrip('https://telegra.ph/')}",
                params=params
        ) as response:
            response = await response.json()

            queue = response['result']['content']
            imgs = []

            while queue:
                curr = queue.pop()

                # print(curr)

                if nexts := curr.get("children"):
                    queue.extend(nexts)

                if img := curr['tag'] == "img":
                    imgs.append(img['attrs']['src'])
                    print(img)

            print(imgs)


            # urls = [filename.split('/')[-1] for filename in
            #         [src['attrs']['src'] for src in response['result']['content'] if 'attrs' in src]]
            # print(f"Images in telegraph page: {len(urls)}") if parser.parse_args().explicit else None
            # if parser.parse_args().mode == 'fast':
            #     await asyncio.gather(*[download_image(url, parser.parse_args().folder, image_id) for image_id, url in enumerate(urls)])
            # else:
            #     [await download_image(url, parser.parse_args().folder, image_id) for image_id, url in enumerate(urls)]


if __name__ == '__main__':
    start_time = datetime.now()
    print(f"Started at: {datetime.now()}")
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
    print(f"Time elapsed: {datetime.now() - start_time}")
