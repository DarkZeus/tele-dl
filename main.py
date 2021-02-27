import asyncio
import pathlib

import ujson
from datetime import datetime
from utils import getsize, convert_bytes, arguments

import aiofiles
import aiohttp


async def download_file(_url, folder, file_id=None):
    if not pathlib.Path(f"{folder}/{file_id}_{_url}").exists() or getsize(f"{folder}/{file_id}_{_url}")['raw'] == 0:
        async with semaphore:
            async with aiohttp.ClientSession(json_serialize=ujson.dumps,
                                             headers={'Connection': 'keep-alive'}) as session:
                async with session.get(f"https://telegra.ph/file/{_url}") as response:
                    if response.status == 200:
                        if not pathlib.Path(folder).exists():
                            try:
                                pathlib.Path(folder).mkdir(parents=True, exist_ok=True)
                            except OSError:
                                print(
                                    f"~> Creation of the directory {folder} failed"
                                ) if parser.parse_args().explicit else None
                            else:
                                print(
                                    f"~> Successfully created the directory {folder}"
                                ) if parser.parse_args().explicit else None

                        path = pathlib.Path().joinpath(f"{folder}/{file_id}_{_url}")
                        async with aiofiles.open(path, 'wb+') as file:
                            await file.write(await response.read())
                            print(
                                f"~> {file_id}_{_url} — {getsize(path)['formatted']}"
                            ) if parser.parse_args().explicit else None
                            await file.flush()


async def main():
    async with aiohttp.ClientSession(json_serialize=ujson.dumps, headers={'Connection': 'keep-alive'}) as session:
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

            await asyncio.gather(*[download_file(
                url,
                parser.parse_args().folder,
                file_id
            ) for file_id, url in enumerate(urls)])

            size = convert_bytes(getsize(parser.parse_args().folder)['raw'] - old_size)
            print(f"~> Saved {size} to {parser.parse_args().folder}",
                  f"~> Time elapsed: {datetime.now() - start_time}",
                  sep="\n")


if __name__ == '__main__':
    parser = arguments()
    semaphore = asyncio.Semaphore(50)
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
