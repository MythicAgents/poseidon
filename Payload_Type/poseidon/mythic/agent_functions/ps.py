from mythic_payloadtype_container.MythicCommandBase import *
import json


class PsArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        if len(self.command_line) > 0:
            if self.command_line[0] == "{":
                tmp_json = json.loads(self.command_line)
                self.command_line = tmp_json["regex_filter"]
            self.add_arg("regex_filter", self.command_line)
        else:
            self.add_arg("regex_filter", "")


class PsCommand(CommandBase):
    cmd = "ps"
    needs_admin = False
    help_cmd = "ps"
    description = "Get a process listing"
    version = 1
    is_exit = False
    is_file_browse = False
    is_process_list = True
    is_download_file = False
    is_remove_file = False
    is_upload_file = False
    author = "@xorrior, @djhohnstein, @its_a_feature"
    argument_class = PsArguments
    attackmapping = ["T1057"]
    browser_script = BrowserScript(script_name="ps", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
