import argparse
import asyncio
import pathlib
from datetime import datetime
from utils import getsize, convert_bytes

import aiofiles
import aiohttp

parser = argparse.ArgumentParser()
parser.add_argument('--link', '-L', help='Enter the full link to the page. Example: '
                                         '"https://telegra.ph/What-Was-TON-And-Why-It-Is-Over-05-12"', type=str,
                    required=True)
parser.add_argument('--folder', '-F', help='Specify the folder where to extract images', type=pathlib.Path,
                    default=pathlib.Path().absolute())
parser.add_argument('--explicit', '-E', help='Show all messages', action="store_true")
parser.add_argument('--mode', '-M', help='Choose mode to download.',
                    choices=['ordered', 'fast'], default='ordered')


async def download_file(_url, folder, file_id=None):
    async with aiohttp.ClientSession() as session:
        async with session.get(f"https://telegra.ph/file/{_url}") as response:
            if response.status == 200:
                if not pathlib.Path(folder).exists():
                    try:
                        pathlib.Path(folder).mkdir(parents=True, exist_ok=True)
                    except OSError:
                        print(f"~> Creation of the directory {folder} failed") if parser.parse_args().explicit else None
                    else:
                        print(
                            f"~> Successfully created the directory {folder}"
                        ) if parser.parse_args().explicit else None

                path = {
                    'fast': pathlib.Path().joinpath(folder, _url),
                    'ordered': pathlib.Path().joinpath(f"{folder}/{file_id}_{_url}"),
                }[parser.parse_args().mode]

                async with aiofiles.open(path, 'wb') as file:
                    await file.write(await response.read())
                    print(f"~> {_url} — {getsize(path)['formatted']}") if parser.parse_args().explicit else None
                    await file.flush()


async def main():
    async with aiohttp.ClientSession() as session:
        async with session.get(
                f"https://api.telegra.ph/getPage/{parser.parse_args().link.removeprefix('https://telegra.ph/')}",
                params={'return_content': 'true'}
        ) as response:
            response = await response.json()

            old_size = getsize(parser.parse_args().folder)['raw']
            start_time = datetime.now()
            print(f"~> Started at: {datetime.now()}",
                  f"~> Saving: {response['result']['title']}",
                  sep="\n")

            queue = response['result']['content']
            files = []

            while queue:
                curr = queue.pop()

                if "children" in curr and (nexts := curr["children"]) and isinstance(nexts, list):
                    queue.extend(nexts)

                if isinstance(curr, dict) and (curr["tag"] == "img" or curr["tag"] == "video"):
                    files.append(curr['attrs']['src'])

            urls = [filename.split('/')[-1] for filename in files[::-1]]
            print(f"~> Files in telegraph page: {len(urls)}") if parser.parse_args().explicit else None

            if parser.parse_args().mode == 'fast':
                await asyncio.gather(*[download_file(
                    url,
                    parser.parse_args().folder,
                    file_id
                ) for file_id, url in enumerate(urls)])
            else:
                [await download_file(
                    url,
                    parser.parse_args().folder,
                    file_id
                ) for file_id, url in enumerate(urls)]

            size = convert_bytes(getsize(parser.parse_args().folder)['raw'] - old_size)
            print(f"~> Saved {size} to {parser.parse_args().folder}",
                  f"~> Time elapsed: {datetime.now() - start_time}",
                  sep="\n")


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
