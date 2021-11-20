import asyncio
from concurrent.futures import ProcessPoolExecutor
from datetime import datetime
from pathlib import Path
from sys import platform

import aiofiles
import aiohttp
import ujson

from parser import arguments
from processing.image import compress_image
from utils import getsize, convert_bytes, append_extension, log, is_image_by_url, create_directory


async def write_file(_bytes: bytes, path: Path):
    async with aiofiles.open(path, 'wb+') as file:
        await file.write(_bytes)
        await file.flush()


async def download_file(_url, folder, file_id=None) -> None:
    create_directory(folder)
    if not Path(f"{folder}/{file_id}_{_url}").exists() or getsize(Path(f"{folder}/{file_id}_{_url}"))['raw'] == 0:
        async with semaphore, aiohttp.ClientSession(json_serialize=ujson.dumps,
                                         headers={'Connection': 'keep-alive'}) as session:
            async with session.get(f"https://telegra.ph/file/{_url}") as response:
                assert response.status == 200

                path = Path(f"{folder}/{file_id}_{_url}")
                if parser.parse_args().compress:
                    loop = asyncio.get_running_loop()
                    executor = ProcessPoolExecutor()
                    destination = append_extension(str(path), "webp")

                    if not Path(destination).exists() or getsize(destination)['formatted'] == 0:
                        if is_image_by_url(_url):
                            await asyncio.gather(*[loop.run_in_executor(
                                executor,
                                compress_image,
                                await response.read(),
                                Path(destination)) if is_image_by_url(_url) else None])
                        else:
                            await write_file(await response.read(), path)
                            log(f"[download] {file_id}_{_url} — {getsize(path)['formatted']}")

                else:
                    await write_file(await response.read(), path)
                    log(f"[download] {file_id}_{_url} — {getsize(path)['formatted']}")


async def main():
    async with aiohttp.ClientSession(json_serialize=ujson.dumps, headers={'Connection': 'keep-alive'}) as session:
        async with session.get(
                f"https://api.telegra.ph/getPage/{parser.parse_args().link.removeprefix('https://telegra.ph/')}",
                params={'return_content': 'true'}
        ) as response:
            response = await response.json()

            old_size = getsize(parser.parse_args().folder)['raw']
            start_time = datetime.now()
            log([
                f"[info] Started at: {datetime.now()}",
                f"[download] {response['result']['title']}",
            ])

            queue = response['result']['content']
            files = []

            while queue:
                curr = queue.pop()

                if "children" in curr and (nexts := curr["children"]) and isinstance(nexts, list):
                    queue.extend(nexts)

                if isinstance(curr, dict) and (curr["tag"] == "img" or curr["tag"] == "video"):
                    files.append(curr['attrs']['src'])

            urls = [filename.split('/')[-1] for filename in files[::-1]]
            log(f"[info] Files in telegraph page: {len(urls)}")

            await asyncio.gather(*[download_file(
                url,
                parser.parse_args().folder,
                file_id
            ) for file_id, url in enumerate(urls)])

            size = convert_bytes(getsize(parser.parse_args().folder)['raw'] - old_size)
            log([
                f"[download] Saved {size} to \"{parser.parse_args().folder}\"",
                f"[info] Time elapsed: {datetime.now() - start_time}"
            ])


if __name__ == '__main__':
    parser = arguments()
    semaphore = asyncio.Semaphore(50)
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy()) if platform == 'win32' else None
    asyncio.run(main())
