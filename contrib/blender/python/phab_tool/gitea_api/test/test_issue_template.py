"""
    Gitea API.

    This documentation describes the Gitea API.  # noqa: E501

    The version of the OpenAPI document: 1.19.0+dev-367-g8042ec1b9
    Generated by: https://openapi-generator.tech
"""


import sys
import unittest

import phab_tool.gitea_api
from phab_tool.gitea_api.model.issue_form_field import IssueFormField
from phab_tool.gitea_api.model.issue_template_labels import IssueTemplateLabels
globals()['IssueFormField'] = IssueFormField
globals()['IssueTemplateLabels'] = IssueTemplateLabels
from phab_tool.gitea_api.model.issue_template import IssueTemplate


class TestIssueTemplate(unittest.TestCase):
    """IssueTemplate unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def testIssueTemplate(self):
        """Test IssueTemplate"""
        # FIXME: construct object with mandatory attributes with example values
        # model = IssueTemplate()  # noqa: E501
        pass


if __name__ == '__main__':
    unittest.main()
