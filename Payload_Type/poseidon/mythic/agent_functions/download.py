from mythic_payloadtype_container.MythicCommandBase import *
import json


class DownloadArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must provide path to thing to download")
        try:
            # if we get JSON, it's from the file browser, so adjust accordingly
            tmp_json = json.loads(self.command_line)
            self.command_line = tmp_json["path"] + "/" + tmp_json["file"]
        except:
            # if it wasn't JSON, then just process it like a normal command-line argument
            pass


class DownloadCommand(CommandBase):
    cmd = "download"
    needs_admin = False
    help_cmd = "download /remote/path/to/file"
    description = "Download a file from the target."
    version = 1
    supported_ui_features = ["file_browser:download"]
    author = "@xorrior"
    argument_class = DownloadArguments
    attackmapping = ["T1022", "T1030", "T1041"]
    browser_script = BrowserScript(script_name="download", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        # adjust the display params to reflect the non-JSON version if needed
        task.display_params = task.args.command_line
        return task

    async def process_response(self, response: AgentResponse):
        pass
