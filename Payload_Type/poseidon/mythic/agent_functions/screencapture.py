from mythic_payloadtype_container.MythicCommandBase import *
import json


class ScreencaptureArguments(TaskArguments):
    def __init__(self, command_line):
        super().__init__(command_line)
        self.args = {}

    async def parse_arguments(self):
        pass


class ScreencaptureCommand(CommandBase):
    cmd = "screencapture"
    needs_admin = False
    help_cmd = "screencapture"
    description = (
        "Capture a screenshot of the targets desktop (not implemented on Linux)."
    )
    version = 1
    author = "@xorrior"
    argument_class = ScreencaptureArguments
    attackmapping = ["T1113"]
    attributes = CommandAttributes(
        # uncomment this when poseidon supports dynamic compilation
        supported_os=[SupportedOS.MacOS]
    )
    browser_script = BrowserScript(script_name="screencapture", author="@djhohnstein")

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
