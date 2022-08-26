from mythic_payloadtype_container.MythicCommandBase import *
import json
from mythic_payloadtype_container.MythicRPC import *


class RunArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = [
            CommandParameter(
                name="path",
                cli_name="path",
                display_name="BinaryPath",
                type=ParameterType.String,
                description="Absolute path to the program to run",
                parameter_group_info=[ParameterGroupInfo(ui_position=1)],
            ),
            CommandParameter(
                name="args",
                cli_name="args",
                display_name="Arguments",
                type=ParameterType.Array,
                description="Array of arguments to pass to the program",
                parameter_group_info=[
                    ParameterGroupInfo(ui_position=2, required=False)
                ],
            ),
        ]

    async def parse_arguments(self):
        self.load_args_from_json_string(self.command_line)
        if self.get_arg("args") is None:
            self.add_arg("args", [])


class RunCommand(CommandBase):
    cmd = "run"
    needs_admin = False
    help_cmd = "run -path /path/to/binary -args arg1 -args arg2 -args arg3"
    description = "Execute a command from disk with arguments"
    version = 1
    author = "@its_a_feature_"
    argument_class = RunArguments
    attackmapping = ["T1059.004"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        resp = await MythicRPC().execute(
            "create_artifact",
            task_id=task.id,
            artifact="{}".format(task.args.command_line),
            artifact_type="Process Create",
        )
        task.display_params = (
            task.args.get_arg("path") + " " + " ".join(task.args.get_arg("args"))
        )
        return task

    async def process_response(self, response: AgentResponse):
        pass
